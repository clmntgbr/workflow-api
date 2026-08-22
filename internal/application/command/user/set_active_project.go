package user

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type SetActiveProjectCommand struct {
	UserID         uuid.UUID
	ProjectID uuid.UUID
}

type SetActiveProjectHandler struct {
	userRepo domainuser.UserWriteRepository
	projectRepo domainproject.ProjectWriteRepository
	outbox   port.OutboxRepository
}

func NewSetActiveProjectHandler(
	userRepo domainuser.UserWriteRepository,
	projectRepo domainproject.ProjectWriteRepository,
	outbox port.OutboxRepository,
) *SetActiveProjectHandler {
	return &SetActiveProjectHandler{
		userRepo: userRepo,
		projectRepo: projectRepo,
		outbox:   outbox,
	}
}

func (h *SetActiveProjectHandler) Handle(ctx context.Context, cmd SetActiveProjectCommand) error {
	if cmd.UserID == uuid.Nil || cmd.ProjectID == uuid.Nil {
		return errors.New("userId and projectId are required")
	}

	return h.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.projectRepo.GetByID(txCtx, cmd.ProjectID)
		if err != nil {
			return errors.New("failed to get project")
		}
		if org == nil {
			return errors.New("project not found")
		}

		isMember := false
		for _, memberID := range org.MemberIDs {
			if memberID == cmd.UserID {
				isMember = true
				break
			}
		}
		if !isMember {
			return errors.New("user is not a member of the project")
		}

		user, err := h.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			return errors.New("failed to get user")
		}
		if user == nil {
			return errors.New("user not found")
		}

		if !user.SetActiveProject(cmd.ProjectID) {
			return nil
		}

		if err := h.userRepo.Update(txCtx, user); err != nil {
			return errors.New("failed to set active project")
		}
		return h.outbox.StoreEvents(txCtx, user.PullEvents())
	})
}
