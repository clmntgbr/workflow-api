package workflowrun

import (
	"context"
	"errors"
	"time"

	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type GetWorkflowRunAnalyticsQuery struct {
	OrganizationID uuid.UUID
	WorkflowID     *uuid.UUID
	From           *time.Time
	To             *time.Time
}

type GetWorkflowRunAnalyticsHandler struct {
	readRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewGetWorkflowRunAnalyticsHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
) *GetWorkflowRunAnalyticsHandler {
	return &GetWorkflowRunAnalyticsHandler{readRepo: readRepo}
}

func (h *GetWorkflowRunAnalyticsHandler) Handle(
	ctx context.Context,
	q GetWorkflowRunAnalyticsQuery,
) (*domainworkflowrun.WorkflowRunAnalytics, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}
	if q.From != nil && q.To != nil && q.From.After(*q.To) {
		return nil, errors.New("from must be before to")
	}

	stats, err := h.readRepo.FindAnalyticsByOrganization(ctx, q.OrganizationID, domainworkflowrun.WorkflowRunAnalyticsFilter{
		WorkflowID: q.WorkflowID,
		From:       q.From,
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
