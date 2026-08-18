package variable

import (
	"context"
	"errors"

	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type DeleteVariableCommand struct {
	ID         uuid.UUID
	WorkflowID uuid.UUID
}

type VariableInUseError struct {
	Steps []domainstep.StepView
}

func (e *VariableInUseError) Error() string {
	return domainvariable.ErrInUse.Error()
}

func (e *VariableInUseError) Unwrap() error {
	return domainvariable.ErrInUse
}

type DeleteVariableHandler struct {
	variableRepo domainvariable.VariableWriteRepository
	stepReadRepo domainstep.StepReadRepository
}

func NewDeleteVariableHandler(
	variableRepo domainvariable.VariableWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
) *DeleteVariableHandler {
	return &DeleteVariableHandler{
		variableRepo: variableRepo,
		stepReadRepo: stepReadRepo,
	}
}

func (h *DeleteVariableHandler) Handle(ctx context.Context, cmd DeleteVariableCommand) error {
	variable, err := h.variableRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get variable")
	}
	if variable == nil || variable.WorkflowID != cmd.WorkflowID {
		return domainvariable.ErrNotFound
	}

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return errors.New("failed to list steps")
	}

	usedBy := make([]domainstep.StepView, 0)
	for _, step := range steps {
		if step.Status == domainstep.StatusDeleted {
			continue
		}
		if referencesVariable(step, variable.Key) {
			usedBy = append(usedBy, step)
		}
	}
	if len(usedBy) > 0 {
		return &VariableInUseError{Steps: usedBy}
	}

	if err := h.variableRepo.Delete(ctx, cmd.ID); err != nil {
		return errors.New("failed to delete variable")
	}
	return nil
}

func referencesVariable(step domainstep.StepView, key string) bool {
	for _, referenced := range domainvariable.CollectReferencedKeys(step.URL, step.Headers, step.Query, step.Body) {
		if referenced == key {
			return true
		}
	}
	return false
}
