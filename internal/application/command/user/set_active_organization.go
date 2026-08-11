package user

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type SetActiveOrganizationCommand struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
}

type SetActiveOrganizationHandler struct {
	userRepo domainuser.UserWriteRepository
	orgRepo  domainorganization.OrganizationWriteRepository
	outbox   port.OutboxRepository
}

func NewSetActiveOrganizationHandler(
	userRepo domainuser.UserWriteRepository,
	orgRepo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *SetActiveOrganizationHandler {
	return &SetActiveOrganizationHandler{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		outbox:   outbox,
	}
}

func (h *SetActiveOrganizationHandler) Handle(ctx context.Context, cmd SetActiveOrganizationCommand) error {
	if cmd.UserID == uuid.Nil || cmd.OrganizationID == uuid.Nil {
		return errors.New("userId and organizationId are required")
	}

	return h.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.orgRepo.GetByID(txCtx, cmd.OrganizationID)
		if err != nil {
			return errors.New("failed to get organization")
		}
		if org == nil {
			return errors.New("organization not found")
		}

		isMember := false
		for _, memberID := range org.MemberIDs {
			if memberID == cmd.UserID {
				isMember = true
				break
			}
		}
		if !isMember {
			return errors.New("user is not a member of the organization")
		}

		user, err := h.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			return errors.New("failed to get user")
		}
		if user == nil {
			return errors.New("user not found")
		}

		if !user.SetActiveOrganization(cmd.OrganizationID) {
			return nil
		}

		if err := h.userRepo.Update(txCtx, user); err != nil {
			return errors.New("failed to set active organization")
		}
		return h.outbox.StoreEvents(txCtx, user.PullEvents())
	})
}
