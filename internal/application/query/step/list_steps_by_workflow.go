package step

import (
	"context"
	"errors"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type ListStepsByWorkflowQuery struct {
	WorkflowID uuid.UUID
}

type ListStepsByWorkflowHandler struct {
	readRepo domainstep.StepReadRepository
}

func NewListStepsByWorkflowHandler(readRepo domainstep.StepReadRepository) *ListStepsByWorkflowHandler {
	return &ListStepsByWorkflowHandler{readRepo: readRepo}
}

func (h *ListStepsByWorkflowHandler) Handle(
	ctx context.Context,
	q ListStepsByWorkflowQuery,
) ([]domainstep.StepView, error) {
	if q.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	views, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list steps")
	}
	return views, nil
}
