package user

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"
)

type DeleteUserByExternalIDHandler struct {
	repo   domainuser.UserWriteRepository
	outbox port.OutboxRepository
}

func NewDeleteUserByExternalIDHandler(
	repo domainuser.UserWriteRepository,
	outbox port.OutboxRepository,
) *DeleteUserByExternalIDHandler {
	return &DeleteUserByExternalIDHandler{repo: repo, outbox: outbox}
}

func (h *DeleteUserByExternalIDHandler) Handle(ctx context.Context, externalID string) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		user, err := h.repo.GetByClerkID(txCtx, externalID)
		if err != nil {
			return errors.New("failed to get user by external ID")
		}
		if user == nil {
			return nil
		}

		user.MarkDeleted()
		events := user.PullEvents()

		if err := h.repo.Delete(txCtx, user.ID); err != nil {
			return errors.New("failed to delete user by external ID")
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
