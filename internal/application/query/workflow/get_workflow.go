package workflow

import (
	"context"
	"errors"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type GetWorkflowByIDQuery struct {
	ID uuid.UUID
}

type GetWorkflowByIDHandler struct {
	readRepo domainworkflow.WorkflowReadRepository
}

func NewGetWorkflowByIDHandler(readRepo domainworkflow.WorkflowReadRepository) *GetWorkflowByIDHandler {
	return &GetWorkflowByIDHandler{readRepo: readRepo}
}

func (h *GetWorkflowByIDHandler) Handle(
	ctx context.Context,
	q GetWorkflowByIDQuery,
) (*domainworkflow.WorkflowView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get workflow")
	}
	if view == nil || view.Status == domainworkflow.StatusDeleted {
		return nil, errors.New("workflow not found")
	}
	return view, nil
}
