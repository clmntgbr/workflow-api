package variable

import (
	"context"
	"errors"
	"strings"

	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type CreateVariableCommand struct {
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
	StepID         uuid.UUID
	Name           string
	Key            string
	Description    string
	Path           string
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
	if err := validateVariableInput(cmd.Name, cmd.Key, cmd.Path); err != nil {
		return nil, err
	}

	step, err := h.stepRepo.GetByID(ctx, cmd.StepID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if step == nil || step.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if step.WorkflowID != cmd.WorkflowID || step.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("step not found")
	}

	existing, err := h.variableReadRepo.FindByWorkflowAndKey(ctx, cmd.WorkflowID, strings.TrimSpace(cmd.Key))
	if err != nil {
		return nil, errors.New("failed to check existing variable")
	}
	if existing != nil {
		return nil, errors.New("variable key already exists in this workflow")
	}

	variable := domainvariable.NewVariable(domainvariable.NewVariableParams{
		Name:           strings.TrimSpace(cmd.Name),
		Key:            strings.TrimSpace(cmd.Key),
		Description:    strings.TrimSpace(cmd.Description),
		Path:           domainvariable.NormalizeJSONPath(strings.TrimSpace(cmd.Path)),
		StepID:         cmd.StepID,
		WorkflowID:     cmd.WorkflowID,
		OrganizationID: cmd.OrganizationID,
	})

	err = h.variableRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.variableRepo.Save(txCtx, variable); err != nil {
			if errors.Is(err, domainvariable.ErrDuplicateKey) {
				return err
			}
			return errors.New("failed to create variable")
		}
		return h.outbox.StoreEvents(txCtx, variable.PullEvents())
	})
	if err != nil {
		return nil, err
	}
	return variable, nil
}

func validateVariableInput(name, key, path string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(key) == "" {
		return errors.New("key is required")
	}
	if strings.TrimSpace(path) == "" {
		return errors.New("path is required")
	}
	return nil
}
