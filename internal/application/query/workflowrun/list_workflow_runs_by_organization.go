package workflowrun

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type ListWorkflowRunsByOrganizationQuery struct {
	OrganizationID uuid.UUID
	Query          paginate.PaginateQuery
}

type ListWorkflowRunsByOrganizationHandler struct {
	readRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewListWorkflowRunsByOrganizationHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
) *ListWorkflowRunsByOrganizationHandler {
	return &ListWorkflowRunsByOrganizationHandler{readRepo: readRepo}
}

func (h *ListWorkflowRunsByOrganizationHandler) Handle(
	ctx context.Context,
	q ListWorkflowRunsByOrganizationQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, 0, errors.New("organizationId is required")
	}
	views, total, err := h.readRepo.FindByOrganizationID(ctx, q.OrganizationID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list workflow runs")
	}
	return views, total, nil
}
