package steprun

import (
	"context"

	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type ListStepRunsByWorkflowRunIDsQuery struct {
	WorkflowRunIDs []uuid.UUID
}

type ListStepRunsByWorkflowRunIDsHandler struct {
	readRepo domainsteprun.StepRunReadRepository
}

func NewListStepRunsByWorkflowRunIDsHandler(
	readRepo domainsteprun.StepRunReadRepository,
) *ListStepRunsByWorkflowRunIDsHandler {
	return &ListStepRunsByWorkflowRunIDsHandler{readRepo: readRepo}
}

func (h *ListStepRunsByWorkflowRunIDsHandler) Handle(
	ctx context.Context,
	q ListStepRunsByWorkflowRunIDsQuery,
) ([]domainsteprun.StepRunView, error) {
	views, err := h.readRepo.FindByWorkflowRunIDs(ctx, q.WorkflowRunIDs)
	if err != nil {
		return nil, err
	}
	return views, nil
}
