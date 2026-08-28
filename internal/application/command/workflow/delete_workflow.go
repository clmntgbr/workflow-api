package workflow

import (
	"context"
	"errors"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type DeleteWorkflowCommand struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type DeleteWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewDeleteWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *DeleteWorkflowHandler {
	return &DeleteWorkflowHandler{repo: repo, outbox: outbox}
}

func (h *DeleteWorkflowHandler) Handle(ctx context.Context, cmd DeleteWorkflowCommand) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		w, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get workflow")
		}
		if w == nil || w.Status == domainworkflow.StatusDeleted {
			return nil
		}

		w.MarkDeleted()

		if err := h.repo.Update(txCtx, w); err != nil {
			return errors.New("failed to delete workflow")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(w.PullEvents(), cmd.UserID))
	})
}
