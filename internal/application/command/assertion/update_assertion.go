package assertion

import (
	"context"
	"errors"
	"strings"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type UpdateAssertionCommand struct {
	ID            uuid.UUID
	WorkflowID    uuid.UUID
	ProjectID     uuid.UUID
	Name          string
	Description   string
	Source        domainassertion.AssertionSource
	Path          string
	Operator      domainassertion.AssertionOperator
	ExpectedValue string
}

type UpdateAssertionHandler struct {
	assertionRepo domainassertion.AssertionWriteRepository
	outbox        port.OutboxRepository
}

func NewUpdateAssertionHandler(
	assertionRepo domainassertion.AssertionWriteRepository,
	outbox port.OutboxRepository,
) *UpdateAssertionHandler {
	return &UpdateAssertionHandler{assertionRepo: assertionRepo, outbox: outbox}
}

func (h *UpdateAssertionHandler) Handle(
	ctx context.Context,
	cmd UpdateAssertionCommand,
) (*domainassertion.Assertion, error) {
	assertion, err := h.assertionRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get assertion")
	}
	if assertion == nil || assertion.WorkflowID != cmd.WorkflowID {
		return nil, domainassertion.ErrNotFound
	}

	if err := assertion.Update(domainassertion.UpdateAssertionParams{
		Name:          strings.TrimSpace(cmd.Name),
		Description:   strings.TrimSpace(cmd.Description),
		Source:        cmd.Source,
		Path:          cmd.Path,
		Operator:      cmd.Operator,
		ExpectedValue: cmd.ExpectedValue,
		ProjectID:     cmd.ProjectID,
	}); err != nil {
		return nil, err
	}

	err = h.assertionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.assertionRepo.Update(txCtx, assertion); err != nil {
			return errors.New("failed to update assertion")
		}
		return h.outbox.StoreEvents(txCtx, assertion.PullEvents())
	})
	if err != nil {
		return nil, err
	}
	return assertion, nil
}
