package handler

import (
	"context"

	usercmd "go-api/internal/application/command/user"
	queryuser "go-api/internal/application/query/user"
	domainuser "go-api/internal/domain/user"
)

type userGetByIDHandler interface {
	Handle(ctx context.Context, q queryuser.GetUserByIDQuery) (*domainuser.UserView, error)
}

type userSetActiveProjectHandler interface {
	Handle(ctx context.Context, cmd usercmd.SetActiveProjectCommand) error
}
