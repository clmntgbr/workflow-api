package activitylog

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/paginate"
	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
)

type ListByWorkflowQuery struct {
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	Query      paginate.PaginateQuery
}

type ListByWorkflowHandler struct {
	readRepo  domainactivitylog.ReadRepository
	retention querysubscription.RunHistoryCutoffResolver
}

func NewListByWorkflowHandler(
	readRepo domainactivitylog.ReadRepository,
	retention querysubscription.RunHistoryCutoffResolver,
) *ListByWorkflowHandler {
	return &ListByWorkflowHandler{readRepo: readRepo, retention: retention}
}

func (h *ListByWorkflowHandler) Handle(
	ctx context.Context,
	q ListByWorkflowQuery,
) ([]domainactivitylog.View, int64, error) {
	if q.WorkflowID == uuid.Nil {
		return nil, 0, errors.New("workflowId is required")
	}

	cutoff, err := h.retention.RunHistoryCutoff(ctx, q.UserID, q.ProjectID)
	if err != nil {
		return nil, 0, errors.New("failed to resolve run history retention")
	}

	views, total, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID, q.Query, cutoff)
	if err != nil {
		return nil, 0, errors.New("failed to list activity logs")
	}
	return views, total, nil
}
