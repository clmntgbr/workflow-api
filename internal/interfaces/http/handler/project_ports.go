package handler

import (
	"context"

	projectcmd "go-api/internal/application/command/project"
	usercmd "go-api/internal/application/command/user"
	queryproject "go-api/internal/application/query/project"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type projectCreateHandler interface {
	Handle(ctx context.Context, cmd projectcmd.CreateProjectCommand) (*domainproject.Project, error)
}

type projectUpdateHandler interface {
	Handle(ctx context.Context, cmd projectcmd.UpdateProjectCommand) error
}

type projectDeleteHandler interface {
	Handle(ctx context.Context, id uuid.UUID) error
}

type projectRemoveMemberHandler interface {
	Handle(ctx context.Context, cmd projectcmd.RemoveProjectMemberCommand) error
}

type projectGetByIDHandler interface {
	Handle(ctx context.Context, q queryproject.GetProjectByIDQuery) (*domainproject.ProjectView, error)
}

type projectListByUserHandler interface {
	Handle(ctx context.Context, q queryproject.ListProjectsByUserQuery) ([]domainproject.ProjectView, int64, error)
}

type projectSetActiveHandler interface {
	Handle(ctx context.Context, cmd usercmd.SetActiveProjectCommand) error
}
