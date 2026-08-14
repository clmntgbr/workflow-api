package variable

import (
	"context"
	"errors"
	"strings"

	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type UpdateVariableCommand struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Key            string
	Description    string
	Path           string
}

type UpdateVariableHandler struct {
	variableRepo domainvariable.VariableWriteRepository
}

func NewUpdateVariableHandler(variableRepo domainvariable.VariableWriteRepository) *UpdateVariableHandler {
	return &UpdateVariableHandler{variableRepo: variableRepo}
}

func (h *UpdateVariableHandler) Handle(ctx context.Context, cmd UpdateVariableCommand) (*domainvariable.Variable, error) {
	if err := validateVariableInput(cmd.Name, cmd.Key, cmd.Path); err != nil {
		return nil, err
	}

	variable, err := h.variableRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get variable")
	}
	if variable == nil || variable.WorkflowID != cmd.WorkflowID {
		return nil, domainvariable.ErrNotFound
	}

	variable.Update(domainvariable.UpdateVariableParams{
		Name:        strings.TrimSpace(cmd.Name),
		Key:         strings.TrimSpace(cmd.Key),
		Description: strings.TrimSpace(cmd.Description),
		Path:        strings.TrimSpace(cmd.Path),
	})

	if err := h.variableRepo.Update(ctx, variable); err != nil {
		if errors.Is(err, domainvariable.ErrDuplicateKey) {
			return nil, err
		}
		return nil, errors.New("failed to update variable")
	}
	return variable, nil
}
