package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type UpdateProjectCommand struct {
	ID   uuid.UUID
	Name string
}

type UpdateProjectHandler struct {
	repo   domainproject.ProjectWriteRepository
	outbox port.OutboxRepository
}

func NewUpdateProjectHandler(
	repo domainproject.ProjectWriteRepository,
	outbox port.OutboxRepository,
) *UpdateProjectHandler {
	return &UpdateProjectHandler{repo: repo, outbox: outbox}
}

func (h *UpdateProjectHandler) Handle(ctx context.Context, cmd UpdateProjectCommand) error {
	if cmd.Name == "" {
		return errors.New("name is required")
	}

	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get project")
		}
		if org == nil {
			return errors.New("project not found")
		}

		org.ApplyUpdate(cmd.Name)

		if err := h.repo.Update(txCtx, org); err != nil {
			return errors.New("failed to update project")
		}
		return h.outbox.StoreEvents(txCtx, org.PullEvents())
	})
}
