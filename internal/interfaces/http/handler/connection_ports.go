package handler

import (
	"context"

	conncmd "go-api/internal/application/command/connection"
	queryconn "go-api/internal/application/query/connection"
	queryworkflow "go-api/internal/application/query/workflow"
	domainconnection "go-api/internal/domain/connection"
	domainworkflow "go-api/internal/domain/workflow"
)

type connectionCreateHandler interface {
	Handle(ctx context.Context, cmd conncmd.CreateConnectionCommand) (*domainconnection.Connection, error)
}

type connectionDeleteHandler interface {
	Handle(ctx context.Context, cmd conncmd.DeleteConnectionCommand) error
}

type connectionListByWorkflowHandler interface {
	Handle(ctx context.Context, q queryconn.ListConnectionsByWorkflowQuery) ([]domainconnection.ConnectionView, error)
}

type connectionGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
