package handler

import (
	"context"

	workflowcmd "go-api/internal/application/command/workflow"
	queryproject "go-api/internal/application/query/project"
	queryworkflow "go-api/internal/application/query/workflow"
	domainproject "go-api/internal/domain/project"
	domainworkflow "go-api/internal/domain/workflow"
)

type workflowCreateHandler interface {
	Handle(ctx context.Context, cmd workflowcmd.CreateWorkflowCommand) (*domainworkflow.Workflow, error)
}

type workflowUpdateHandler interface {
	Handle(ctx context.Context, cmd workflowcmd.UpdateWorkflowCommand) error
}

type workflowActivateHandler interface {
	Handle(ctx context.Context, cmd workflowcmd.ActivateWorkflowCommand) (*domainworkflow.Workflow, error)
}

type workflowDeactivateHandler interface {
	Handle(ctx context.Context, cmd workflowcmd.DeactivateWorkflowCommand) (*domainworkflow.Workflow, error)
}

type workflowDeleteHandler interface {
	Handle(ctx context.Context, cmd workflowcmd.DeleteWorkflowCommand) error
}

type workflowGetByIDHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}

type workflowListByProjectHandler interface {
	Handle(ctx context.Context, q queryworkflow.ListWorkflowsByProjectQuery) ([]domainworkflow.WorkflowView, int64, error)
}

type workflowGetProjectByIDHandler interface {
	Handle(ctx context.Context, q queryproject.GetProjectByIDQuery) (*domainproject.ProjectView, error)
}
