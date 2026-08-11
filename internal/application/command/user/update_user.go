package user

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"
)

type UpdateUserCommand struct {
	User      *domainuser.User
	FirstName string
	LastName  string
	Email     string
	Banned    bool
}

type UpdateUserHandler struct {
	repo   domainuser.UserWriteRepository
	outbox port.OutboxRepository
}

func NewUpdateUserHandler(repo domainuser.UserWriteRepository, outbox port.OutboxRepository) *UpdateUserHandler {
	return &UpdateUserHandler{repo: repo, outbox: outbox}
}

func (h *UpdateUserHandler) Handle(ctx context.Context, cmd UpdateUserCommand) error {
	if cmd.User == nil {
		return errors.New("user is required")
	}

	cmd.User.ApplyUpdate(cmd.FirstName, cmd.LastName, cmd.Email, cmd.Banned)

	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Update(txCtx, cmd.User); err != nil {
			return errors.New("failed to update user")
		}
		return h.outbox.StoreEvents(txCtx, cmd.User.PullEvents())
	})
}
