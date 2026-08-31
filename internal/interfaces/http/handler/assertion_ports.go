package handler

import (
	"context"

	assertioncmd "go-api/internal/application/command/assertion"
	queryassertion "go-api/internal/application/query/assertion"
	querystep "go-api/internal/application/query/step"
	queryworkflow "go-api/internal/application/query/workflow"
	domainassertion "go-api/internal/domain/assertion"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"
)

type assertionCreateHandler interface {
	Handle(ctx context.Context, cmd assertioncmd.CreateAssertionCommand) (*domainassertion.Assertion, error)
}

type assertionUpdateHandler interface {
	Handle(ctx context.Context, cmd assertioncmd.UpdateAssertionCommand) (*domainassertion.Assertion, error)
}

type assertionDeleteHandler interface {
	Handle(ctx context.Context, cmd assertioncmd.DeleteAssertionCommand) error
}

type assertionGetByIDHandler interface {
	Handle(ctx context.Context, q queryassertion.GetAssertionByIDQuery) (*domainassertion.AssertionView, error)
}

type assertionListByStepHandler interface {
	Handle(ctx context.Context, q queryassertion.ListAssertionsByStepQuery) ([]domainassertion.AssertionView, error)
}

type assertionSearchPathsHandler interface {
	Handle(ctx context.Context, q queryassertion.SearchAssertionPathsQuery) ([]string, int, error)
}

type assertionGetStepHandler interface {
	Handle(ctx context.Context, q querystep.GetStepByIDQuery) (*domainstep.StepView, error)
}

type assertionGetWorkflowHandler interface {
	Handle(ctx context.Context, q queryworkflow.GetWorkflowByIDQuery) (*domainworkflow.WorkflowView, error)
}
