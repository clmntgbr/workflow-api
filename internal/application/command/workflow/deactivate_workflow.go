package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type DeactivateWorkflowCommand struct {
	ID             uuid.UUID
	ProjectID uuid.UUID
}

type DeactivateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewDeactivateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *DeactivateWorkflowHandler {
	return &DeactivateWorkflowHandler{repo: repo, outbox: outbox}
}

func (h *DeactivateWorkflowHandler) Handle(ctx context.Context, cmd DeactivateWorkflowCommand) (*domainworkflow.Workflow, error) {
	var workflow *domainworkflow.Workflow
	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		w, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get workflow")
		}
		if w == nil || w.Status == domainworkflow.StatusDeleted || w.ProjectID != cmd.ProjectID {
			return errors.New("workflow not found")
		}
		if err := w.Deactivate(); err != nil {
			return err
		}
		if err := h.repo.Update(txCtx, w); err != nil {
			return errors.New("failed to deactivate workflow")
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
