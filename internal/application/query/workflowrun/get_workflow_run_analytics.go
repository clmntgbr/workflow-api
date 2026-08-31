package workflowrun

import (
	"context"
	"errors"
	"time"

	querysubscription "go-api/internal/application/query/subscription"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type GetWorkflowRunAnalyticsQuery struct {
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	From       *time.Time
	To         *time.Time
}

type GetWorkflowRunAnalyticsHandler struct {
	readRepo  domainworkflowrun.WorkflowRunReadRepository
	retention querysubscription.RunHistoryCutoffResolver
}

func NewGetWorkflowRunAnalyticsHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
	retention querysubscription.RunHistoryCutoffResolver,
) *GetWorkflowRunAnalyticsHandler {
	return &GetWorkflowRunAnalyticsHandler{readRepo: readRepo, retention: retention}
}

func (h *GetWorkflowRunAnalyticsHandler) Handle(
	ctx context.Context,
	q GetWorkflowRunAnalyticsQuery,
) (*domainworkflowrun.WorkflowRunAnalytics, error) {
	if q.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if q.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	if q.From != nil && q.To != nil && q.From.After(*q.To) {
		return nil, errors.New("from must be before to")
	}

	cutoff, err := h.retention.RunHistoryCutoff(ctx, q.UserID, q.ProjectID)
	if err != nil {
		return nil, errors.New("failed to resolve run history retention")
	}
	from := querysubscription.ClampTimeFrom(q.From, cutoff)

	workflowID := q.WorkflowID
	stats, err := h.readRepo.FindAnalyticsByProject(ctx, q.ProjectID, domainworkflowrun.WorkflowRunAnalyticsFilter{
		WorkflowID: &workflowID,
		From:       from,
		To:         q.To,
	})
	if err != nil {
		return nil, errors.New("failed to load workflow run analytics")
	}
	if stats.TotalRuns == 0 {
		return stats, nil
	}

	total := float64(stats.TotalRuns)
	successRate := (float64(stats.SuccessCount) / total) * 100
	failureRate := (float64(stats.FailureCount) / total) * 100
	stats.SuccessRate = successRate
	stats.FailureRate = failureRate
	return stats, nil
}
