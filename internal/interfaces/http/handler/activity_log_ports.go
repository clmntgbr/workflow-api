package handler

import (
	"context"

	queryactivitylog "go-api/internal/application/query/activitylog"
	queryworkflow "go-api/internal/application/query/workflow"
	domainactivitylog "go-api/internal/domain/activitylog"
	domainworkflow "go-api/internal/domain/workflow"
)

type activityLogListByWorkflowHandler interface {
	Handle(ctx context.Context, q queryactivitylog.ListByWorkflowQuery) ([]domainactivitylog.View, int64, error)
}

type activityLogGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
