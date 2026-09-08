package workflowrun

import (
	"context"
	"errors"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/event"
	domaininsight "go-api/internal/domain/insight"
	"go-api/internal/domain/port"
	domainquota "go-api/internal/domain/quota"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type StartWorkflowRunCommand struct {
	WorkflowID              uuid.UUID
	TriggeredBy             domainworkflowrun.TriggeredBy
	TriggeredByUserID       *uuid.UUID
	Context                 map[string]any
	FromStepID              *uuid.UUID
	ScheduleAlreadyAdvanced bool
}

type StartWorkflowRunHandler struct {
	workflowRepo domainworkflow.WorkflowWriteRepository
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	variableRead domainvariable.VariableReadRepository
	stepRead     domainstep.StepReadRepository
	connRead     domainconnection.ConnectionReadRepository
	stepRunRead  domainsteprun.StepRunReadRepository
	stepRunWrite domainsteprun.StepRunWriteRepository
	insightRead  domaininsight.InsightReadRepository
	insightWrite domaininsight.InsightWriteRepository
	outbox       port.OutboxRepository
	assert       *cmdquota.AssertCreateAllowedHandler
}

func NewStartWorkflowRunHandler(
	workflowRepo domainworkflow.WorkflowWriteRepository,
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	variableRead domainvariable.VariableReadRepository,
	stepRead domainstep.StepReadRepository,
	connRead domainconnection.ConnectionReadRepository,
	stepRunRead domainsteprun.StepRunReadRepository,
	stepRunWrite domainsteprun.StepRunWriteRepository,
	insightRead domaininsight.InsightReadRepository,
	insightWrite domaininsight.InsightWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *StartWorkflowRunHandler {
	return &StartWorkflowRunHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		variableRead: variableRead,
		stepRead:     stepRead,
		connRead:     connRead,
		stepRunRead:  stepRunRead,
		stepRunWrite: stepRunWrite,
		insightRead:  insightRead,
		insightWrite: insightWrite,
		outbox:       outbox,
		assert:       assert,
	}
}

func (h *StartWorkflowRunHandler) Handle(
	ctx context.Context,
	cmd StartWorkflowRunCommand,
) (*domainworkflowrun.WorkflowRun, error) {
	if cmd.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	if !cmd.TriggeredBy.Valid() {
		return nil, errors.New("invalid triggeredBy")
	}

	workflow, err := h.workflowRepo.GetByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to start workflow run")
	}
	if workflow == nil || workflow.Status == domainworkflow.StatusDeleted {
		return nil, domainworkflowrun.ErrWorkflowNotFound
	}

	inProgress, err := h.runRepo.HasInProgress(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to start workflow run")
	}
	if inProgress {
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule {
			if err := h.recordScheduledSkip(
				ctx,
				workflow,
				cmd.ScheduleAlreadyAdvanced,
				domainworkflowrun.ScheduledSkipReasonAlreadyInProgress,
				false,
			); err != nil {
				return nil, errors.New("failed to start workflow run")
			}
		}
		return nil, domainworkflowrun.ErrAlreadyInProgress
	}

	if err := h.assert.AssertWorkflowRunStart(
		ctx,
		workflow.ProjectID,
		cmd.TriggeredByUserID,
		1,
	); err != nil {
		quotaName := domainquota.QuotaNameWorkflowRuns
		if errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded) {
			quotaName = domainquota.QuotaNameConcurrentRuns
		}
		if errors.Is(err, cmdquota.ErrWorkflowRunQuotaExceeded) || errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded) {
			_ = h.storeQuotaExceeded(ctx, workflow, cmd.TriggeredByUserID, quotaName)
		}
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule {
			switch {
			case errors.Is(err, cmdquota.ErrWorkflowRunQuotaExceeded):
				if skipErr := h.recordScheduledSkip(
					ctx,
					workflow,
					cmd.ScheduleAlreadyAdvanced,
					domainworkflowrun.ScheduledSkipReasonQuotaExceeded,
					true,
				); skipErr != nil {
					return nil, errors.New("failed to start workflow run")
				}
				return nil, err
			case errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded):
				if skipErr := h.recordScheduledSkip(
					ctx,
					workflow,
					cmd.ScheduleAlreadyAdvanced,
					domainworkflowrun.ScheduledSkipReasonQuotaExceeded,
					false,
				); skipErr != nil {
					return nil, errors.New("failed to start workflow run")
				}
				return nil, err
			}
		}
		return nil, err
	}

	variables, err := h.variableRead.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to start workflow run")
	}
	runContext := domainvariable.SeedContextWithStaticVariables(cmd.Context, variables)

	run := domainworkflowrun.NewWorkflowRun(domainworkflowrun.NewWorkflowRunParams{
		WorkflowID:        cmd.WorkflowID,
		TriggeredBy:       cmd.TriggeredBy,
		TriggeredByUserID: cmd.TriggeredByUserID,
		Context:           runContext,
	})

	var replay *startFromReplay
	if cmd.FromStepID != nil && *cmd.FromStepID != uuid.Nil {
		replay, err = h.prepareStartFrom(ctx, *cmd.FromStepID, run)
		if err != nil {
			if errors.Is(err, domainworkflowrun.ErrFromStepNotFound) ||
				errors.Is(err, domainworkflowrun.ErrMissingPreviousStepRun) {
				return nil, err
			}
			return nil, errors.New("failed to start workflow run")
		}
	}

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.runRepo.Save(txCtx, run); err != nil {
			return err
		}
		if replay != nil {
			for _, copied := range replay.copies {
				if err := h.stepRunWrite.Save(txCtx, copied); err != nil {
					return err
				}
			}
			for _, skipped := range replay.skips {
				if err := h.stepRunWrite.Save(txCtx, skipped); err != nil {
					return err
				}
			}
			for _, insight := range replay.insights {
				if err := h.insightWrite.Save(txCtx, insight); err != nil {
					return err
				}
			}
		}
		events := run.PullEvents()
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule && !cmd.ScheduleAlreadyAdvanced {
			workflow.AdvanceAfterScheduledStart(time.Now().UTC())
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
			events = append(events, workflow.PullEvents()...)
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		if isUniqueViolation(err) {
			if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule {
				if skipErr := h.recordScheduledSkip(
					ctx,
					workflow,
					cmd.ScheduleAlreadyAdvanced,
					domainworkflowrun.ScheduledSkipReasonAlreadyInProgress,
					false,
				); skipErr != nil {
					return nil, errors.New("failed to start workflow run")
				}
			}
			return nil, domainworkflowrun.ErrAlreadyInProgress
		}
		return nil, errors.New("failed to start workflow run")
	}

	_ = h.storeQuotaThresholdIfNeeded(ctx, workflow, cmd.TriggeredByUserID)
	return run, nil
}

func (h *StartWorkflowRunHandler) storeQuotaThresholdIfNeeded(
	ctx context.Context,
	workflow *domainworkflow.Workflow,
	preferredUserID *uuid.UUID,
) error {
	snap, err := h.assert.WorkflowRunQuotaSnapshot(ctx, workflow.ProjectID, preferredUserID)
	if err != nil || snap == nil || snap.Max <= 0 {
		return err
	}
	usedAfter := snap.Used + 1
	if usedAfter*100 < int64(snap.Max)*int64(cmdquota.WorkflowRunWarningPercent) {
		return nil
	}
	percent := int(usedAfter * 100 / int64(snap.Max))
	now := time.Now().UTC()
	evt := domainquota.QuotaThresholdReached{
		ID: event.DeterministicID(
			domainquota.EventTypeQuotaThresholdReached,
			snap.SubscriptionID.String(),
			snap.PeriodKey(),
			domainquota.QuotaNameWorkflowRuns,
			"80",
		),
		SubscriptionID: snap.SubscriptionID.String(),
		UserID:         snap.UserID.String(),
		QuotaName:      domainquota.QuotaNameWorkflowRuns,
		Used:           usedAfter,
		Max:            int64(snap.Max),
		Percent:        percent,
		Timestamp:      now,
	}
	return h.outbox.StoreEvents(ctx, []event.DomainEvent{evt})
}

func (h *StartWorkflowRunHandler) storeQuotaExceeded(
	ctx context.Context,
	workflow *domainworkflow.Workflow,
	preferredUserID *uuid.UUID,
	quotaName string,
) error {
	snap, err := h.assert.WorkflowRunQuotaSnapshot(ctx, workflow.ProjectID, preferredUserID)
	if err != nil || snap == nil {
		return err
	}
	now := time.Now().UTC()
	evt := domainquota.QuotaExceeded{
		ID: event.DeterministicID(
			domainquota.EventTypeQuotaExceeded,
			snap.SubscriptionID.String(),
			snap.PeriodKey(),
			quotaName,
		),
		SubscriptionID: snap.SubscriptionID.String(),
		UserID:         snap.UserID.String(),
		QuotaName:      quotaName,
		Used:           snap.Used,
		Max:            int64(snap.Max),
		Timestamp:      now,
	}
	return h.outbox.StoreEvents(ctx, []event.DomainEvent{evt})
}

func (h *StartWorkflowRunHandler) recordScheduledSkip(
	ctx context.Context,
	workflow *domainworkflow.Workflow,
	scheduleAlreadyAdvanced bool,
	reason string,
	clearSchedule bool,
) error {
	now := time.Now().UTC()
	skipped := domainworkflowrun.WorkflowRunScheduledSkipped{
		ID:         uuid.New().String(),
		WorkflowID: workflow.ID.String(),
		Reason:     reason,
		Timestamp:  now,
	}

	return h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		events := []event.DomainEvent{skipped}

		if clearSchedule {
			workflow.ClearSchedule()
			events = append(events, workflow.PullEvents()...)
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
		} else if !scheduleAlreadyAdvanced {
			workflow.AdvanceAfterScheduledStart(now)
			events = append(events, workflow.PullEvents()...)
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
