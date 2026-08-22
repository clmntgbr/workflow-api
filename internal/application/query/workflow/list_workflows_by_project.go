package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ListWorkflowsByProjectQuery struct {
	ProjectID uuid.UUID
	Query          paginate.PaginateQuery
}

type ListWorkflowsByProjectHandler struct {
	readRepo domainworkflow.WorkflowReadRepository
}

func NewListWorkflowsByProjectHandler(
	readRepo domainworkflow.WorkflowReadRepository,
) *ListWorkflowsByProjectHandler {
	return &ListWorkflowsByProjectHandler{readRepo: readRepo}
}

func (h *ListWorkflowsByProjectHandler) Handle(
	ctx context.Context,
	q ListWorkflowsByProjectQuery,
) ([]domainworkflow.WorkflowView, int64, error) {
	if q.ProjectID == uuid.Nil {
		return nil, 0, errors.New("projectId is required")
	}
	views, total, err := h.readRepo.FindByProjectID(ctx, q.ProjectID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list workflows")
	}
	return views, total, nil
}
