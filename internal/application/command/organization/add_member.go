package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type AddOrganizationMemberCommand struct {
	OrganizationID uuid.UUID
	UserID         uuid.UUID
}

type AddOrganizationMemberHandler struct {
	orgRepo  domainorganization.OrganizationWriteRepository
	userRepo domainuser.UserWriteRepository
	outbox   port.OutboxRepository
}

func NewAddOrganizationMemberHandler(
	orgRepo domainorganization.OrganizationWriteRepository,
	userRepo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *AddOrganizationMemberHandler {
	return &AddOrganizationMemberHandler{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		outbox:   outbox,
	}
}

func (h *AddOrganizationMemberHandler) Handle(ctx context.Context, cmd AddOrganizationMemberCommand) error {
	return h.orgRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.orgRepo.GetByID(txCtx, cmd.OrganizationID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		if !org.AddMember(cmd.UserID) {
			return nil
		}

		if err := h.orgRepo.Update(txCtx, org); err != nil {
			return errors.New("failed to add organization member")
		}

		events := org.PullEvents()

		user, err := h.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			return errors.New("failed to get user")
		}
		if user != nil && user.ActiveOrganizationID == nil {
			user.SetActiveOrganization(org.ID)
			if err := h.userRepo.Update(txCtx, user); err != nil {
				return errors.New("failed to set active organization")
			}
			events = append(events, user.PullEvents()...)
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
