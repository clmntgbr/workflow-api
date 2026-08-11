package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type DeleteOrganizationHandler struct {
	repo   domainorganization.OrganizationWriteRepository
	outbox port.OutboxRepository
}

func NewDeleteOrganizationHandler(
	repo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *DeleteOrganizationHandler {
	return &DeleteOrganizationHandler{repo: repo, outbox: outbox}
}

func (h *DeleteOrganizationHandler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, id)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return nil
		}

		org.MarkDeleted()
		events := org.PullEvents()

		if err := h.repo.Delete(txCtx, org.ID); err != nil {
			return errors.New("failed to delete organization")
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
