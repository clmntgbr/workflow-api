package workflowrun

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type ListWorkflowRunsByProjectQuery struct {
	ProjectID uuid.UUID
	Query          paginate.PaginateQuery
}

type ListWorkflowRunsByProjectHandler struct {
	readRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewListWorkflowRunsByProjectHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
) *ListWorkflowRunsByProjectHandler {
	return &ListWorkflowRunsByProjectHandler{readRepo: readRepo}
}

func (h *ListWorkflowRunsByProjectHandler) Handle(
	ctx context.Context,
	q ListWorkflowRunsByProjectQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	if q.ProjectID == uuid.Nil {
		return nil, 0, errors.New("projectId is required")
	}
	views, total, err := h.readRepo.FindByProjectID(ctx, q.ProjectID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list workflow runs")
	}
	return views, total, nil
}
