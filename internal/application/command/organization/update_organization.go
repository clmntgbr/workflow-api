package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type UpdateOrganizationCommand struct {
	ID   uuid.UUID
	Name string
}

type UpdateOrganizationHandler struct {
	repo   domainorganization.OrganizationWriteRepository
	outbox port.OutboxRepository
}

func NewUpdateOrganizationHandler(
	repo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *UpdateOrganizationHandler {
	return &UpdateOrganizationHandler{repo: repo, outbox: outbox}
}

func (h *UpdateOrganizationHandler) Handle(ctx context.Context, cmd UpdateOrganizationCommand) error {
	if cmd.Name == "" {
		return errors.New("name is required")
	}

	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		org.ApplyUpdate(cmd.Name)

		if err := h.repo.Update(txCtx, org); err != nil {
			return errors.New("failed to update organization")
		}
		return h.outbox.StoreEvents(txCtx, org.PullEvents())
	})
}
