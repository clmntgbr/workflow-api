package workflowrun

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type StartWorkflowRunCommand struct {
	WorkflowID        uuid.UUID
	TriggeredBy       domainworkflowrun.TriggeredBy
	TriggeredByUserID *uuid.UUID
	Context           map[string]any
}

type StartWorkflowRunHandler struct {
	workflowRepo domainworkflow.WorkflowWriteRepository
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	outbox       port.OutboxRepository
}

func NewStartWorkflowRunHandler(
	workflowRepo domainworkflow.WorkflowWriteRepository,
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	outbox port.OutboxRepository,
) *StartWorkflowRunHandler {
	return &StartWorkflowRunHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		outbox:       outbox,
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
		return nil, domainworkflowrun.ErrAlreadyInProgress
	}

	run := domainworkflowrun.NewWorkflowRun(domainworkflowrun.NewWorkflowRunParams{
		WorkflowID:        cmd.WorkflowID,
		TriggeredBy:       cmd.TriggeredBy,
		TriggeredByUserID: cmd.TriggeredByUserID,
		Context:           cmd.Context,
	})

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.runRepo.Save(txCtx, run); err != nil {
			return err
		}
		if cmd.TriggeredBy == domainworkflowrun.TriggeredBySchedule {
			workflow.AdvanceAfterScheduledStart(time.Now().UTC())
			if err := h.workflowRepo.Update(txCtx, workflow); err != nil {
				return err
			}
		}
		return h.outbox.StoreEvents(txCtx, run.PullEvents())
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domainworkflowrun.ErrAlreadyInProgress
		}
		return nil, errors.New("failed to start workflow run")
	}

	return run, nil
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
