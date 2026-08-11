package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
)

type CreateOrganizationCommand struct {
	Name string
}

type CreateOrganizationHandler struct {
	repo   domainorganization.OrganizationWriteRepository
	outbox port.OutboxRepository
}

func NewCreateOrganizationHandler(
	repo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *CreateOrganizationHandler {
	return &CreateOrganizationHandler{repo: repo, outbox: outbox}
}

func (h *CreateOrganizationHandler) Handle(
	ctx context.Context,
	cmd CreateOrganizationCommand,
) (*domainorganization.Organization, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}

	org := domainorganization.NewOrganization(cmd.Name)

	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, org); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, org.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create organization")
	}

	return org, nil
}
