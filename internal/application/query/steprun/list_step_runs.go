package steprun

import (
	"context"
	"errors"

	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type ListStepRunsByWorkflowRunQuery struct {
	WorkflowRunID uuid.UUID
}

type ListStepRunsByWorkflowRunHandler struct {
	readRepo domainsteprun.StepRunReadRepository
}

func NewListStepRunsByWorkflowRunHandler(
	readRepo domainsteprun.StepRunReadRepository,
) *ListStepRunsByWorkflowRunHandler {
	return &ListStepRunsByWorkflowRunHandler{readRepo: readRepo}
}

func (h *ListStepRunsByWorkflowRunHandler) Handle(
	ctx context.Context,
	q ListStepRunsByWorkflowRunQuery,
) ([]domainsteprun.StepRunView, error) {
	if q.WorkflowRunID == uuid.Nil {
		return nil, errors.New("workflowRunId is required")
	}
	views, err := h.readRepo.FindByWorkflowRunID(ctx, q.WorkflowRunID)
	if err != nil {
		return nil, errors.New("failed to list step runs")
	}
	return views, nil
}
