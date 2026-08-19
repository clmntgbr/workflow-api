package workflowrun

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type CancelWorkflowRunCommand struct {
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
}

type CancelWorkflowRunHandler struct {
	workflowRepo domainworkflow.WorkflowWriteRepository
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	outbox       port.OutboxRepository
}

func NewCancelWorkflowRunHandler(
	workflowRepo domainworkflow.WorkflowWriteRepository,
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	outbox port.OutboxRepository,
) *CancelWorkflowRunHandler {
	return &CancelWorkflowRunHandler{
		workflowRepo: workflowRepo,
		runRepo:      runRepo,
		outbox:       outbox,
	}
}

func (h *CancelWorkflowRunHandler) Handle(
	ctx context.Context,
	cmd CancelWorkflowRunCommand,
) (*domainworkflowrun.WorkflowRun, error) {
	if cmd.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}

	workflow, err := h.workflowRepo.GetByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to cancel workflow run")
	}
	if workflow == nil ||
		workflow.Status == domainworkflow.StatusDeleted ||
		workflow.OrganizationID != cmd.OrganizationID {
		return nil, domainworkflowrun.ErrWorkflowNotFound
	}

	run, err := h.runRepo.FindInProgressByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to cancel workflow run")
	}
	if run == nil {
		return nil, domainworkflowrun.ErrNoRunInProgress
	}

	if err := run.MarkCancelled(); err != nil {
		if errors.Is(err, domainworkflowrun.ErrAlreadyTerminal) {
			return nil, domainworkflowrun.ErrNoRunInProgress
		}
		return nil, err
	}

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.runRepo.Update(txCtx, run); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, run.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to cancel workflow run")
	}

	return run, nil
}
