package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type RemoveOrganizationMemberCommand struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

type RemoveOrganizationMemberHandler struct {
	orgRepo  domainorganization.OrganizationWriteRepository
	userRepo domainuser.UserWriteRepository
	outbox   port.OutboxRepository
}

func NewRemoveOrganizationMemberHandler(
	orgRepo domainorganization.OrganizationWriteRepository,
	userRepo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *RemoveOrganizationMemberHandler {
	return &RemoveOrganizationMemberHandler{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		outbox:   outbox,
	}
}

func (h *RemoveOrganizationMemberHandler) Handle(ctx context.Context, cmd RemoveOrganizationMemberCommand) error {
	return h.orgRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.orgRepo.GetByID(txCtx, cmd.OrganizationID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		if !org.RemoveMember(cmd.UserID) {
			return nil
		}

		if err := h.orgRepo.Update(txCtx, org); err != nil {
			return errors.New("failed to remove organization member")
		}

		events := org.PullEvents()

		user, err := h.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			return errors.New("failed to get user")
		}
		if user != nil &&
			user.ActiveOrganizationID != nil &&
			*user.ActiveOrganizationID == cmd.OrganizationID {
			user.ClearActiveOrganization()
			if err := h.userRepo.Update(txCtx, user); err != nil {
				return errors.New("failed to clear active organization")
			}
			events = append(events, user.PullEvents()...)
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
