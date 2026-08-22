package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type DeleteProjectHandler struct {
	repo   domainproject.ProjectWriteRepository
	outbox port.OutboxRepository
}

func NewDeleteProjectHandler(
	repo domainproject.ProjectWriteRepository,
	outbox port.OutboxRepository,
) *DeleteProjectHandler {
	return &DeleteProjectHandler{repo: repo, outbox: outbox}
}

func (h *DeleteProjectHandler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		org, err := h.repo.GetByID(txCtx, id)
		if err != nil {
			return errors.New("failed to get project")
		}
		if org == nil {
			return nil
		}

		org.MarkDeleted()
		events := org.PullEvents()

		if err := h.repo.Delete(txCtx, org.ID); err != nil {
			return errors.New("failed to delete project")
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
