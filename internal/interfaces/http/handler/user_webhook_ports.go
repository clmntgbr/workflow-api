package handler

import (
	"context"

	usercmd "go-api/internal/application/command/user"
	domainuser "go-api/internal/domain/user"
)

type userWebhookGetByExternalIDHandler interface {
	Handle(ctx context.Context, externalID string) (*domainuser.User, error)
}

type userWebhookCreateUserHandler interface {
	Handle(ctx context.Context, cmd usercmd.CreateUserCommand) (*domainuser.User, error)
}

type userWebhookUpdateUserHandler interface {
	Handle(ctx context.Context, cmd usercmd.UpdateUserCommand) error
}

type userWebhookDeleteByExternalIDHandler interface {
	Handle(ctx context.Context, externalID string) error
}
