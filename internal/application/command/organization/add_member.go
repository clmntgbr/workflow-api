package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type AddOrganizationMemberCommand struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

type AddOrganizationMemberHandler struct {
	repo   domainorganization.OrganizationWriteRepository
	outbox port.OutboxRepository
}

func NewAddOrganizationMemberHandler(
	repo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *AddOrganizationMemberHandler {
	return &AddOrganizationMemberHandler{repo: repo, outbox: outbox}
}

func (h *AddOrganizationMemberHandler) Handle(ctx context.Context, cmd AddOrganizationMemberCommand) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, cmd.OrganizationID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		if !org.AddMember(cmd.UserID) {
			return nil
		}

		if err := h.repo.Update(txCtx, org); err != nil {
			return errors.New("failed to add organization member")
		}
		return h.outbox.StoreEvents(txCtx, org.PullEvents())
	})
}
