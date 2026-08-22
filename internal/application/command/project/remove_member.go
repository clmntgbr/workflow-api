package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type RemoveProjectMemberCommand struct {
	ProjectID uuid.UUID
	UserID         uuid.UUID
}

type RemoveProjectMemberHandler struct {
	projectRepo domainproject.ProjectWriteRepository
	userRepo domainuser.UserWriteRepository
	outbox   port.OutboxRepository
}

func NewRemoveProjectMemberHandler(
	projectRepo domainproject.ProjectWriteRepository,
	userRepo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *RemoveProjectMemberHandler {
	return &RemoveProjectMemberHandler{
		projectRepo: projectRepo,
		userRepo: userRepo,
		outbox:   outbox,
	}
}

func (h *RemoveProjectMemberHandler) Handle(ctx context.Context, cmd RemoveProjectMemberCommand) error {
	return h.projectRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.projectRepo.GetByID(txCtx, cmd.ProjectID)
		if err != nil {
			return errors.New("failed to get project")
		}
		if org == nil {
			return errors.New("project not found")
		}

		if !org.RemoveMember(cmd.UserID) {
			return nil
		}

		if err := h.projectRepo.Update(txCtx, org); err != nil {
			return errors.New("failed to remove project member")
		}

		events := org.PullEvents()

		user, err := h.userRepo.GetByID(txCtx, cmd.UserID)
		if err != nil {
			return errors.New("failed to get user")
		}
		if user != nil &&
			user.ActiveProjectID != nil &&
			*user.ActiveProjectID == cmd.ProjectID {
			user.ClearActiveProject()
			if err := h.userRepo.Update(txCtx, user); err != nil {
				return errors.New("failed to clear active project")
			}
			events = append(events, user.PullEvents()...)
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
