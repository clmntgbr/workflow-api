package activitylog

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
)

type ListByWorkflowQuery struct {
	WorkflowID uuid.UUID
	Query      paginate.PaginateQuery
}

type ListByWorkflowHandler struct {
	readRepo domainactivitylog.ReadRepository
}

func NewListByWorkflowHandler(readRepo domainactivitylog.ReadRepository) *ListByWorkflowHandler {
	return &ListByWorkflowHandler{readRepo: readRepo}
}

func (h *ListByWorkflowHandler) Handle(
	ctx context.Context,
	q ListByWorkflowQuery,
) ([]domainactivitylog.View, int64, error) {
	if q.WorkflowID == uuid.Nil {
		return nil, 0, errors.New("workflowId is required")
	}
	views, total, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list activity logs")
	}
	return views, total, nil
}
