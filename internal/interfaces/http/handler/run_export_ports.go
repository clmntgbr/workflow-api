package handler

import (
	"context"

	runexportcmd "go-api/internal/application/command/runexport"
	queryrunexport "go-api/internal/application/query/runexport"
	queryworkflow "go-api/internal/application/query/workflow"
	domainrunexport "go-api/internal/domain/runexport"
	domainworkflow "go-api/internal/domain/workflow"
)

type runExportRequestHandler interface {
	Handle(ctx context.Context, cmd runexportcmd.RequestRunExportCommand) (*domainrunexport.RunExport, error)
}

type runExportGetByIDHandler interface {
	Handle(ctx context.Context, q queryrunexport.GetRunExportByIDQuery) (*domainrunexport.View, error)
}

type runExportGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
