package assertion

import (
	"context"
	"errors"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
)

type DeleteAssertionCommand struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	WorkflowID uuid.UUID
	ProjectID  uuid.UUID
}

type DeleteAssertionHandler struct {
	assertionRepo domainassertion.AssertionWriteRepository
	outbox        port.OutboxRepository
}

func NewDeleteAssertionHandler(
	assertionRepo domainassertion.AssertionWriteRepository,
	outbox port.OutboxRepository,
) *DeleteAssertionHandler {
	return &DeleteAssertionHandler{
		assertionRepo: assertionRepo,
		outbox:        outbox,
	}
}

func (h *DeleteAssertionHandler) Handle(ctx context.Context, cmd DeleteAssertionCommand) error {
	assertion, err := h.assertionRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get assertion")
	}
	if assertion == nil || assertion.WorkflowID != cmd.WorkflowID {
		return domainassertion.ErrNotFound
	}

	assertion.MarkDeleted(cmd.ProjectID)
	return h.assertionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.assertionRepo.Delete(txCtx, cmd.ID); err != nil {
			return errors.New("failed to delete assertion")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(assertion.PullEvents(), cmd.UserID))
	})
}
