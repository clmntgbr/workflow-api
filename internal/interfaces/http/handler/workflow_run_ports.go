package handler

import (
	"context"

	workflowruncmd "go-api/internal/application/command/workflowrun"
	queryinsight "go-api/internal/application/query/insight"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	domaininsight "go-api/internal/domain/insight"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"
)

type workflowRunStartHandler interface {
	Handle(ctx context.Context, cmd workflowruncmd.StartWorkflowRunCommand) (*domainworkflowrun.WorkflowRun, error)
}

type workflowRunCancelHandler interface {
	Handle(ctx context.Context, cmd workflowruncmd.CancelWorkflowRunCommand) (*domainworkflowrun.WorkflowRun, error)
}

type workflowRunGetByIDHandler interface {
	Handle(ctx context.Context, q queryworkflowrun.GetWorkflowRunByIDQuery) (*domainworkflowrun.WorkflowRunView, error)
}

type workflowRunAnalyticsHandler interface {
	Handle(ctx context.Context, q queryworkflowrun.GetWorkflowRunAnalyticsQuery) (*domainworkflowrun.WorkflowRunAnalytics, error)
}

type workflowRunListByWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflowrun.ListWorkflowRunsByWorkflowQuery) ([]domainworkflowrun.WorkflowRunView, int64, error)
}

type workflowRunListStepRunsHandler interface {
	Handle(ctx context.Context, q querysteprun.ListStepRunsByWorkflowRunQuery) ([]domainsteprun.StepRunView, error)
}

type workflowRunListStepRunsByIDsHandler interface {
	Handle(ctx context.Context, q querysteprun.ListStepRunsByWorkflowRunIDsQuery) ([]domainsteprun.StepRunView, error)
}

type workflowRunListInsightsByIDsHandler interface {
	Handle(ctx context.Context, q queryinsight.ListInsightsByStepRunIDsQuery) ([]domaininsight.InsightView, error)
}

type workflowRunListStepsByWorkflowHandler interface {
	Handle(ctx context.Context, q querystep.ListStepsByWorkflowQuery) ([]domainstep.StepView, error)
}
