package variable

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	"go-api/internal/application/messaging"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type DeleteVariableCommand struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	WorkflowID uuid.UUID
	ProjectID  uuid.UUID
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
	outbox       port.OutboxRepository
}

func NewDeleteVariableHandler(
	variableRepo domainvariable.VariableWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	outbox port.OutboxRepository,
) *DeleteVariableHandler {
	return &DeleteVariableHandler{
		variableRepo: variableRepo,
		stepReadRepo: stepReadRepo,
		outbox:       outbox,
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

	variable.MarkDeleted(cmd.ProjectID)
	return h.variableRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.variableRepo.Delete(txCtx, cmd.ID); err != nil {
			return errors.New("failed to delete variable")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(variable.PullEvents(), cmd.UserID))
	})
}

func referencesVariable(step domainstep.StepView, key string) bool {
	for _, referenced := range domainvariable.CollectReferencedKeys(step.URL, step.Headers, step.Query, step.Body) {
		if referenced == key {
			return true
		}
	}
	return false
}
