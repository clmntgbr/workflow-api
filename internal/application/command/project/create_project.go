package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type CreateProjectCommand struct {
	Name          string
	CreatorUserID uuid.UUID
}

type CreateProjectHandler struct {
	projectRepo domainproject.ProjectWriteRepository
	userRepo domainuser.UserWriteRepository
	outbox   port.OutboxRepository
}

func NewCreateProjectHandler(
	projectRepo domainproject.ProjectWriteRepository,
	userRepo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *CreateProjectHandler {
	return &CreateProjectHandler{
		projectRepo: projectRepo,
		userRepo: userRepo,
		outbox:   outbox,
	}
}

func (h *CreateProjectHandler) Handle(
	ctx context.Context,
	cmd CreateProjectCommand,
) (*domainproject.Project, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.CreatorUserID == uuid.Nil {
		return nil, errors.New("creator user is required")
	}

	org := domainproject.NewProject(cmd.Name, cmd.CreatorUserID)
	org.AddMember(cmd.CreatorUserID)

	err := h.projectRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.projectRepo.Save(txCtx, org); err != nil {
			return err
		}

		user, err := h.userRepo.GetByID(txCtx, cmd.CreatorUserID)
		if err != nil {
			return errors.New("failed to get creator user")
		}
		if user == nil {
			return errors.New("creator user not found")
		}

		user.SetActiveProject(org.ID)
		if err := h.userRepo.Update(txCtx, user); err != nil {
			return errors.New("failed to set active project")
		}

		events := append(org.PullEvents(), user.PullEvents()...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		if err.Error() == "creator user not found" || err.Error() == "failed to get creator user" ||
			err.Error() == "failed to set active project" {
			return nil, err
		}
		return nil, errors.New("failed to create project")
	}

	return org, nil
}
