package variable

import (
	"context"
	"errors"

	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type DeleteVariableCommand struct {
	ID         uuid.UUID
	WorkflowID uuid.UUID
}

type DeleteVariableHandler struct {
	variableRepo domainvariable.VariableWriteRepository
}

func NewDeleteVariableHandler(variableRepo domainvariable.VariableWriteRepository) *DeleteVariableHandler {
	return &DeleteVariableHandler{variableRepo: variableRepo}
}

func (h *DeleteVariableHandler) Handle(ctx context.Context, cmd DeleteVariableCommand) error {
	variable, err := h.variableRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get variable")
	}
	if variable == nil || variable.WorkflowID != cmd.WorkflowID {
		return domainvariable.ErrNotFound
	}
	if err := h.variableRepo.Delete(ctx, cmd.ID); err != nil {
		return errors.New("failed to delete variable")
	}
	return nil
}
