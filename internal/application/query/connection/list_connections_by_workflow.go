package connection

import (
	"context"

	domainconnection "go-api/internal/domain/connection"

	"github.com/google/uuid"
)

type ListConnectionsByWorkflowQuery struct {
	WorkflowID uuid.UUID
}

type ListConnectionsByWorkflowHandler struct {
	repo domainconnection.ConnectionReadRepository
}

func NewListConnectionsByWorkflowHandler(
	repo domainconnection.ConnectionReadRepository,
) *ListConnectionsByWorkflowHandler {
	return &ListConnectionsByWorkflowHandler{repo: repo}
}

func (h *ListConnectionsByWorkflowHandler) Handle(
	ctx context.Context,
	q ListConnectionsByWorkflowQuery,
) ([]domainconnection.ConnectionView, error) {
	return h.repo.FindByWorkflowID(ctx, q.WorkflowID)
}
