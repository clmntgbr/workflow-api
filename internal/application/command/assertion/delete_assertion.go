package assertion

import (
	"context"
	"errors"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
)

type DeleteAssertionCommand struct {
	ID         uuid.UUID
	WorkflowID uuid.UUID
}

type DeleteAssertionHandler struct {
	assertionRepo domainassertion.AssertionWriteRepository
}

func NewDeleteAssertionHandler(
	assertionRepo domainassertion.AssertionWriteRepository,
) *DeleteAssertionHandler {
	return &DeleteAssertionHandler{assertionRepo: assertionRepo}
}

func (h *DeleteAssertionHandler) Handle(ctx context.Context, cmd DeleteAssertionCommand) error {
	assertion, err := h.assertionRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get assertion")
	}
	if assertion == nil || assertion.WorkflowID != cmd.WorkflowID {
		return domainassertion.ErrNotFound
	}
	if err := h.assertionRepo.Delete(ctx, cmd.ID); err != nil {
		return errors.New("failed to delete assertion")
	}
	return nil
}
