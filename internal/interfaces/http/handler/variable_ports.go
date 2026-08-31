package handler

import (
	"context"

	variablecmd "go-api/internal/application/command/variable"
	querystep "go-api/internal/application/query/step"
	queryvariable "go-api/internal/application/query/variable"
	queryworkflow "go-api/internal/application/query/workflow"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
)

type variableCreateHandler interface {
	Handle(ctx context.Context, cmd variablecmd.CreateVariableCommand) (*domainvariable.Variable, error)
}

type variableUpdateHandler interface {
	Handle(ctx context.Context, cmd variablecmd.UpdateVariableCommand) (*domainvariable.Variable, error)
}

type variableDeleteHandler interface {
	Handle(ctx context.Context, cmd variablecmd.DeleteVariableCommand) error
}

type variableGetByIDHandler interface {
	Handle(ctx context.Context, q queryvariable.GetVariableByIDQuery) (*domainvariable.VariableView, error)
}

type variableListByWorkflowHandler interface {
	Handle(ctx context.Context, q queryvariable.ListVariablesByWorkflowQuery) ([]domainvariable.VariableView, error)
}

type variableListAvailableHandler interface {
	Handle(ctx context.Context, q queryvariable.ListAvailableVariablesQuery) ([]domainvariable.VariableView, error)
}

type variableSearchPathsHandler interface {
	Handle(ctx context.Context, q queryvariable.SearchVariablePathsQuery) ([]string, int, error)
}

type variableGetStepHandler interface {
	Handle(ctx context.Context, q querystep.GetStepByIDQuery) (*domainstep.StepView, error)
}

type variableGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
