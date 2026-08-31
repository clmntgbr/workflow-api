package workflowrun

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type ListWorkflowRunsByWorkflowQuery struct {
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	Query      paginate.PaginateQuery
}

type ListWorkflowRunsByWorkflowHandler struct {
	readRepo  domainworkflowrun.WorkflowRunReadRepository
	retention querysubscription.RunHistoryCutoffResolver
}

func NewListWorkflowRunsByWorkflowHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
	retention querysubscription.RunHistoryCutoffResolver,
) *ListWorkflowRunsByWorkflowHandler {
	return &ListWorkflowRunsByWorkflowHandler{readRepo: readRepo, retention: retention}
}

func (h *ListWorkflowRunsByWorkflowHandler) Handle(
	ctx context.Context,
	q ListWorkflowRunsByWorkflowQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	if q.WorkflowID == uuid.Nil {
		return nil, 0, errors.New("workflowId is required")
	}

	cutoff, err := h.retention.RunHistoryCutoff(ctx, q.UserID, q.ProjectID)
	if err != nil {
		return nil, 0, errors.New("failed to resolve run history retention")
	}

	filter := domainworkflowrun.WorkflowRunListFilter{
		Paginate:     q.Query,
		CreatedAfter: cutoff,
	}

	views, total, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID, filter)
	if err != nil {
		return nil, 0, errors.New("failed to list workflow runs")
	}
	return views, total, nil
}
