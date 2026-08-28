package variable

import (
	"context"
	"errors"
	"strings"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type CreateVariableCommand struct {
	UserID         uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
	StepID         *uuid.UUID
	Kind           domainvariable.Kind
	Name           string
	Key            string
	Description    string
	Path           string
	Value          any
}

type CreateVariableHandler struct {
	variableRepo     domainvariable.VariableWriteRepository
	variableReadRepo domainvariable.VariableReadRepository
	stepRepo         domainstep.StepWriteRepository
	outbox           port.OutboxRepository
}

func NewCreateVariableHandler(
	variableRepo domainvariable.VariableWriteRepository,
	variableReadRepo domainvariable.VariableReadRepository,
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *CreateVariableHandler {
	return &CreateVariableHandler{
		variableRepo:     variableRepo,
		variableReadRepo: variableReadRepo,
		stepRepo:         stepRepo,
		outbox:           outbox,
	}
}

func (h *CreateVariableHandler) Handle(ctx context.Context, cmd CreateVariableCommand) (*domainvariable.Variable, error) {
	if err := domainvariable.ValidateIdentity(cmd.Name, cmd.Key); err != nil {
		return nil, err
	}

	kind := cmd.Kind
	if kind == "" {
		if cmd.StepID != nil && *cmd.StepID != uuid.Nil {
			kind = domainvariable.KindExtracted
		} else {
			kind = domainvariable.KindStatic
		}
	}
	if !kind.Valid() {
		return nil, domainvariable.ErrInvalidKind
	}

	var stepID *uuid.UUID
	path := strings.TrimSpace(cmd.Path)
	var value any

	switch kind {
	case domainvariable.KindExtracted:
		if cmd.StepID == nil || *cmd.StepID == uuid.Nil {
			return nil, domainvariable.ErrStepRequired
		}
		step, err := h.stepRepo.GetByID(ctx, *cmd.StepID)
		if err != nil {
			return nil, errors.New("failed to get step")
		}
	if step == nil || step.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if step.WorkflowID != cmd.WorkflowID || step.ProjectID != cmd.ProjectID {
		return nil, errors.New("step not found")
	}
	if step.Type == domainstep.TypeDelay {
		return nil, domainstep.ErrDelayStepCannotHaveExtras
	}
	stepID = cmd.StepID
		value = nil
	case domainvariable.KindStatic:
		stepID = nil
		path = ""
		value = cmd.Value
	}

	existing, err := h.variableReadRepo.FindByWorkflowAndKey(ctx, cmd.WorkflowID, strings.TrimSpace(cmd.Key))
	if err != nil {
		return nil, errors.New("failed to check existing variable")
	}
	if existing != nil {
		return nil, errors.New("variable key already exists in this workflow")
	}

	variable, err := domainvariable.NewVariable(domainvariable.NewVariableParams{
		Name:           strings.TrimSpace(cmd.Name),
		Key:            strings.TrimSpace(cmd.Key),
		Description:    strings.TrimSpace(cmd.Description),
		Kind:           kind,
		Path:           path,
		Value:          value,
		StepID:         stepID,
		WorkflowID:     cmd.WorkflowID,
		ProjectID: cmd.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	err = h.variableRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.variableRepo.Save(txCtx, variable); err != nil {
			if errors.Is(err, domainvariable.ErrDuplicateKey) {
				return err
			}
			return errors.New("failed to create variable")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(variable.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, err
	}
	return variable, nil
}
