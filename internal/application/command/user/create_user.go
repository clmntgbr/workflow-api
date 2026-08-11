package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"
)

type CreateUserCommand struct {
	ClerkID   string
	FirstName string
	LastName  string
	Email     string
	Banned    bool
}

type CreateUserHandler struct {
	userRepo domainuser.UserWriteRepository
	orgRepo  domainorganization.OrganizationWriteRepository
	outbox   port.OutboxRepository
}

func NewCreateUserHandler(
	userRepo domainuser.UserWriteRepository,
	orgRepo domainorganization.OrganizationWriteRepository,
	outbox port.OutboxRepository,
) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo: userRepo,
		orgRepo:  orgRepo,
		outbox:   outbox,
	}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*domainuser.User, error) {
	u := domainuser.NewUser(cmd.ClerkID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Banned)

	err := h.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.userRepo.Save(txCtx, u); err != nil {
			return err
		}

		org := domainorganization.NewOrganization(personalOrganizationName(cmd.FirstName, cmd.LastName), u.ID)
		org.AddMember(u.ID)
		if err := h.orgRepo.Save(txCtx, org); err != nil {
			return err
		}

		u.SetActiveOrganization(org.ID)
		if err := h.userRepo.Update(txCtx, u); err != nil {
			return err
		}

		events := append(u.PullEvents(), org.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return u, nil
}

func personalOrganizationName(firstName, lastName string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	switch {
	case firstName != "" && lastName != "":
		return fmt.Sprintf("%s %s", firstName, lastName)
	case firstName != "":
		return fmt.Sprintf("%s's Organization", firstName)
	default:
		return "Personal Organization"
	}
}
