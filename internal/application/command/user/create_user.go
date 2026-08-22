package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainproject "go-api/internal/domain/project"
	domainplan "go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
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
	userRepo         domainuser.UserWriteRepository
	projectRepo      domainproject.ProjectWriteRepository
	planRepo         domainplan.PlanWriteRepository
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewCreateUserHandler(
	userRepo domainuser.UserWriteRepository,
	projectRepo domainproject.ProjectWriteRepository,
	planRepo domainplan.PlanWriteRepository,
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo:         userRepo,
		projectRepo: projectRepo,
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*domainuser.User, error) {
	u := domainuser.NewUser(cmd.ClerkID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Banned)

	err := h.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.userRepo.Save(txCtx, u); err != nil {
			return err
		}

		org := domainproject.NewProject(personalProjectName(cmd.FirstName, cmd.LastName), u.ID)
		org.AddMember(u.ID)
		if err := h.projectRepo.Save(txCtx, org); err != nil {
			return err
		}

		freePlan, err := h.planRepo.GetBySlug(txCtx, domainplan.FreePlanSlug)
		if err != nil {
			return err
		}
		if freePlan == nil {
			return errors.New("free plan not found")
		}

		sub := domainsubscription.NewFreeSubscription(freePlan.ID)
		if err := h.subscriptionRepo.Save(txCtx, sub); err != nil {
			return err
		}

		u.SetActiveProject(org.ID)
		u.AssignSubscription(sub.ID)
		if err := h.userRepo.Update(txCtx, u); err != nil {
			return err
		}

		events := append(u.PullEvents(), org.PullEvents()...)
		events = append(events, sub.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return u, nil
}

func personalProjectName(firstName, lastName string) string {
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)
	switch {
	case firstName != "" && lastName != "":
		return fmt.Sprintf("%s %s", firstName, lastName)
	case firstName != "":
		return fmt.Sprintf("%s's Project", firstName)
	default:
		return "Personal Project"
	}
}
