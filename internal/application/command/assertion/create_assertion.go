package assertion

import (
	"context"
	"errors"
	"strings"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type CreateAssertionCommand struct {
	WorkflowID    uuid.UUID
	ProjectID     uuid.UUID
	StepID        uuid.UUID
	Name          string
	Description   string
	Source        domainassertion.AssertionSource
	Path          string
	Operator      domainassertion.AssertionOperator
	ExpectedValue string
}

type CreateAssertionHandler struct {
	assertionRepo domainassertion.AssertionWriteRepository
	stepRepo      domainstep.StepWriteRepository
	outbox        port.OutboxRepository
}

func NewCreateAssertionHandler(
	assertionRepo domainassertion.AssertionWriteRepository,
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *CreateAssertionHandler {
	return &CreateAssertionHandler{
		assertionRepo: assertionRepo,
		stepRepo:      stepRepo,
		outbox:        outbox,
	}
}

func (h *CreateAssertionHandler) Handle(
	ctx context.Context,
	cmd CreateAssertionCommand,
) (*domainassertion.Assertion, error) {
	step, err := h.stepRepo.GetByID(ctx, cmd.StepID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if step == nil || step.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if step.WorkflowID != cmd.WorkflowID || step.ProjectID != cmd.ProjectID {
		return nil, errors.New("step not found")
	}

	assertion, err := domainassertion.NewAssertion(domainassertion.NewAssertionParams{
		Name:          strings.TrimSpace(cmd.Name),
		Description:   strings.TrimSpace(cmd.Description),
		Source:        cmd.Source,
		Path:          cmd.Path,
		Operator:      cmd.Operator,
		ExpectedValue: cmd.ExpectedValue,
		StepID:        cmd.StepID,
		WorkflowID:    cmd.WorkflowID,
		ProjectID:     cmd.ProjectID,
	})
	if err != nil {
		return nil, err
	}

	err = h.assertionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.assertionRepo.Save(txCtx, assertion); err != nil {
			return errors.New("failed to create assertion")
		}
		return h.outbox.StoreEvents(txCtx, assertion.PullEvents())
	})
	if err != nil {
		return nil, err
	}
	return assertion, nil
}
