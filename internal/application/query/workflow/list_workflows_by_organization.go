package workflow

import (
	"context"
	"errors"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ListWorkflowsByOrganizationQuery struct {
	OrganizationID uuid.UUID
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
) ([]domainworkflow.WorkflowView, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}
	views, err := h.readRepo.FindByOrganizationID(ctx, q.OrganizationID)
	if err != nil {
		return nil, errors.New("failed to list workflows")
	}
	return views, nil
}
