package handler

import (
	"context"

	stepcmd "go-api/internal/application/command/step"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type stepCreateHandler interface {
	Handle(ctx context.Context, cmd stepcmd.CreateStepCommand) (*domainstep.Step, error)
}

type stepCreateDelayHandler interface {
	Handle(ctx context.Context, cmd stepcmd.CreateDelayStepCommand) (*domainstep.Step, error)
}

type stepCreateConditionHandler interface {
	Handle(ctx context.Context, cmd stepcmd.CreateConditionStepCommand) (*domainstep.Step, error)
}

type stepUpdateHandler interface {
	Handle(ctx context.Context, cmd stepcmd.UpdateStepCommand) (*domainstep.Step, error)
}

type stepUpdateDelayHandler interface {
	Handle(ctx context.Context, cmd stepcmd.UpdateDelayStepCommand) (*domainstep.Step, error)
}

type stepUpdateConditionHandler interface {
	Handle(ctx context.Context, cmd stepcmd.UpdateConditionStepCommand) (*domainstep.Step, error)
}

type stepUpdatePositionHandler interface {
	Handle(ctx context.Context, cmd stepcmd.UpdateStepPositionCommand) (*domainstep.Step, error)
}

type stepDeleteHandler interface {
	Handle(ctx context.Context, cmd stepcmd.DeleteStepCommand) error
}

type stepGetByIDHandler interface {
	Handle(ctx context.Context, q querystep.GetStepByIDQuery) (*domainstep.StepView, error)
}

type stepListByWorkflowHandler interface {
	Handle(ctx context.Context, q querystep.ListStepsByWorkflowQuery) ([]domainstep.StepView, error)
}

type stepLatestRunStatusHandler interface {
	Handle(ctx context.Context, q querysteprun.GetLatestStepRunStatusesByStepIDsQuery) (map[uuid.UUID]domainsteprun.Status, error)
}

type stepGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
