package workflowrun

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type ListWorkflowRunsByWorkflowQuery struct {
	WorkflowID uuid.UUID
	Query      paginate.PaginateQuery
}

type ListWorkflowRunsByWorkflowHandler struct {
	readRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewListWorkflowRunsByWorkflowHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
) *ListWorkflowRunsByWorkflowHandler {
	return &ListWorkflowRunsByWorkflowHandler{readRepo: readRepo}
}

func (h *ListWorkflowRunsByWorkflowHandler) Handle(
	ctx context.Context,
	q ListWorkflowRunsByWorkflowQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	if q.WorkflowID == uuid.Nil {
		return nil, 0, errors.New("workflowId is required")
	}
	views, total, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list workflow runs")
	}
	return views, total, nil
}
