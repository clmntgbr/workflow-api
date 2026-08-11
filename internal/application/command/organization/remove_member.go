package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type RemoveOrganizationMemberCommand struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

type RemoveOrganizationMemberHandler struct {
	repo   domainorganization.OrganizationWriteRepository
	outbox port.OutboxRepository
}

func NewRemoveOrganizationMemberHandler(
	repo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *RemoveOrganizationMemberHandler {
	return &RemoveOrganizationMemberHandler{repo: repo, outbox: outbox}
}

func (h *RemoveOrganizationMemberHandler) Handle(ctx context.Context, cmd RemoveOrganizationMemberCommand) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, cmd.OrganizationID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		if !org.RemoveMember(cmd.UserID) {
			return nil
		}

		if err := h.repo.Update(txCtx, org); err != nil {
			return errors.New("failed to remove organization member")
		}
		return h.outbox.StoreEvents(txCtx, org.PullEvents())
	})
}
