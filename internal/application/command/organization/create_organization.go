package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type CreateOrganizationCommand struct {
	Name          string
	CreatorUserID uuid.UUID
}

type CreateOrganizationHandler struct {
	orgRepo  domainorganization.OrganizationWriteRepository
	userRepo domainuser.UserWriteRepository
	outbox   port.OutboxRepository
}

func NewCreateOrganizationHandler(
	orgRepo domainorganization.OrganizationWriteRepository,
	userRepo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *CreateOrganizationHandler {
	return &CreateOrganizationHandler{
		orgRepo:  orgRepo,
		userRepo: userRepo,
		outbox:   outbox,
	}
}

func (h *CreateOrganizationHandler) Handle(
	ctx context.Context,
	cmd CreateOrganizationCommand,
) (*domainorganization.Organization, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.CreatorUserID == uuid.Nil {
		return nil, errors.New("creator user is required")
	}

	org := domainorganization.NewOrganization(cmd.Name, cmd.CreatorUserID)
	org.AddMember(cmd.CreatorUserID)

	err := h.orgRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.orgRepo.Save(txCtx, org); err != nil {
			return err
		}

		user, err := h.userRepo.GetByID(txCtx, cmd.CreatorUserID)
		if err != nil {
			return errors.New("failed to get creator user")
		}
		if user == nil {
			return errors.New("creator user not found")
		}

		user.SetActiveOrganization(org.ID)
		if err := h.userRepo.Update(txCtx, user); err != nil {
			return errors.New("failed to set active organization")
		}

		events := append(org.PullEvents(), user.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		if err.Error() == "creator user not found" || err.Error() == "failed to get creator user" ||
			err.Error() == "failed to set active organization" {
			return nil, err
		}
		return nil, errors.New("failed to create organization")
	}

	return org, nil
}
