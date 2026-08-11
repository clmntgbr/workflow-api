package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ListWorkflowsByOrganizationQuery struct {
	OrganizationID uuid.UUID
	Query          paginate.PaginateQuery
}

type ListWorkflowsByOrganizationHandler struct {
	readRepo domainworkflow.WorkflowReadRepository
}

func NewListWorkflowsByOrganizationHandler(
	readRepo domainworkflow.WorkflowReadRepository,
) *ListWorkflowsByOrganizationHandler {
	return &ListWorkflowsByOrganizationHandler{readRepo: readRepo}
}

func (h *ListWorkflowsByOrganizationHandler) Handle(
	ctx context.Context,
	q ListWorkflowsByOrganizationQuery,
) ([]domainworkflow.WorkflowView, int64, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, 0, errors.New("organizationId is required")
	}
	views, total, err := h.readRepo.FindByOrganizationID(ctx, q.OrganizationID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list workflows")
	}
	return views, total, nil
}
