package user

import (
	"context"
	"errors"

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
	repo   domainuser.UserWriteRepository
	outbox port.OutboxRepository
}

func NewCreateUserHandler(repo domainuser.UserWriteRepository, outbox port.OutboxRepository) *CreateUserHandler {
	return &CreateUserHandler{repo: repo, outbox: outbox}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*domainuser.User, error) {
	u := domainuser.NewUser(cmd.ClerkID, cmd.FirstName, cmd.LastName, cmd.Email, cmd.Banned)

	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, u); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, u.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	return u, nil
}
