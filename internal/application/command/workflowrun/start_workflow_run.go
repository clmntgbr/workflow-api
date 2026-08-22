package workflowrun

import (
	"context"
	"errors"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
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
	ScheduleAlreadyAdvanced bool
}

type StartWorkflowRunHandler struct {
	workflowRepo domainworkflow.WorkflowWriteRepository
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	variableRead domainvariable.VariableReadRepository
	outbox       port.OutboxRepository
	assert       *cmdquota.AssertCreateAllowedHandler
}

func NewStartWorkflowRunHandler(
	workflowRepo domainworkflow.WorkflowWriteRepository,
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	variableRead domainvariable.VariableReadRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *StartWorkflowRunHandler {
	return &StartWorkflowRunHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		variableRead: variableRead,
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
			); err != nil {
				return nil, errors.New("failed to start workflow run")
			}
		}
		return nil, domainworkflowrun.ErrAlreadyInProgress
	}

	if err := h.assert.AssertWorkflowRunStart(
		ctx,
		workflow.OrganizationID,
		cmd.TriggeredByUserID,
		1,
	); err != nil {
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule && isQuotaExceeded(err) {
			if skipErr := h.recordScheduledSkip(
				ctx,
				workflow,
				cmd.ScheduleAlreadyAdvanced,
				domainworkflowrun.ScheduledSkipReasonQuotaExceeded,
			); skipErr != nil {
				return nil, errors.New("failed to start workflow run")
			}
			return nil, err
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

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.runRepo.Save(txCtx, run); err != nil {
			return err
		}
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule && !cmd.ScheduleAlreadyAdvanced {
			workflow.AdvanceAfterScheduledStart(time.Now().UTC())
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
		}
		return h.outbox.StoreEvents(txCtx, run.PullEvents())
	})
	if err != nil {
		if isUniqueViolation(err) {
			if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule {
				if skipErr := h.recordScheduledSkip(
					ctx,
					workflow,
					cmd.ScheduleAlreadyAdvanced,
					domainworkflowrun.ScheduledSkipReasonAlreadyInProgress,
				); skipErr != nil {
					return nil, errors.New("failed to start workflow run")
				}
			}
			return nil, domainworkflowrun.ErrAlreadyInProgress
		}
		return nil, errors.New("failed to start workflow run")
	}

	return run, nil
}

func isQuotaExceeded(err error) bool {
	return errors.Is(err, cmdquota.ErrWorkflowRunQuotaExceeded) ||
		errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded)
}

func (h *StartWorkflowRunHandler) recordScheduledSkip(
	ctx context.Context,
	workflow *domainworkflow.Workflow,
	scheduleAlreadyAdvanced bool,
	reason string,
) error {
	now := time.Now().UTC()
	skipped := domainworkflowrun.WorkflowRunScheduledSkipped{
		ID:         uuid.New().String(),
		WorkflowID: workflow.ID.String(),
		Reason:     reason,
		Timestamp:  now,
	}

	return h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if !scheduleAlreadyAdvanced {
			workflow.AdvanceAfterScheduledStart(now)
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
		}
		return h.outbox.StoreEvents(txCtx, []event.DomainEvent{skipped})
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
