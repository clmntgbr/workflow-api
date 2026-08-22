package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ActivateWorkflowCommand struct {
	ID             uuid.UUID
	ProjectID uuid.UUID
}

type ActivateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewActivateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *ActivateWorkflowHandler {
	return &ActivateWorkflowHandler{repo: repo, outbox: outbox}
}

func (h *ActivateWorkflowHandler) Handle(ctx context.Context, cmd ActivateWorkflowCommand) (*domainworkflow.Workflow, error) {
	var workflow *domainworkflow.Workflow
	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		w, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get workflow")
		}
		if w == nil || w.Status == domainworkflow.StatusDeleted || w.ProjectID != cmd.ProjectID {
			return errors.New("workflow not found")
		}
		if err := w.Activate(); err != nil {
			return err
		}
		if err := h.repo.Update(txCtx, w); err != nil {
			return errors.New("failed to activate workflow")
		}
		if err := h.outbox.StoreEvents(txCtx, w.PullEvents()); err != nil {
			return err
		}
		workflow = w
		return nil
	})
	if err != nil {
		return nil, err
	}
	return workflow, nil
}
