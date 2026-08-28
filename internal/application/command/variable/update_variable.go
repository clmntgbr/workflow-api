package variable

import (
	"context"
	"errors"
	"strings"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type UpdateVariableCommand struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
	Name           string
	Key            string
	Description    string
	Path           string
	Value          any
}

type UpdateVariableHandler struct {
	variableRepo domainvariable.VariableWriteRepository
	outbox       port.OutboxRepository
}

func NewUpdateVariableHandler(
	variableRepo domainvariable.VariableWriteRepository,
	outbox port.OutboxRepository,
) *UpdateVariableHandler {
	return &UpdateVariableHandler{variableRepo: variableRepo, outbox: outbox}
}

func (h *UpdateVariableHandler) Handle(ctx context.Context, cmd UpdateVariableCommand) (*domainvariable.Variable, error) {
	if err := domainvariable.ValidateIdentity(cmd.Name, cmd.Key); err != nil {
		return nil, err
	}

	variable, err := h.variableRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get variable")
	}
	if variable == nil || variable.WorkflowID != cmd.WorkflowID {
		return nil, domainvariable.ErrNotFound
	}

	path := strings.TrimSpace(cmd.Path)
	value := cmd.Value
	if variable.Kind == domainvariable.KindStatic {
		path = ""
	} else {
		value = nil
	}

	if err := variable.Update(domainvariable.UpdateVariableParams{
		Name:           strings.TrimSpace(cmd.Name),
		Key:            strings.TrimSpace(cmd.Key),
		Description:    strings.TrimSpace(cmd.Description),
		Path:           path,
		Value:          value,
		ProjectID: cmd.ProjectID,
	}); err != nil {
		return nil, err
	}

	err = h.variableRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.variableRepo.Update(txCtx, variable); err != nil {
			if errors.Is(err, domainvariable.ErrDuplicateKey) {
				return err
			}
			return errors.New("failed to update variable")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(variable.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, err
	}
	return variable, nil
}
