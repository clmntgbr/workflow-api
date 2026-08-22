package workflowrun

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type WorkflowRunWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, run *WorkflowRun) error
	Update(ctx context.Context, run *WorkflowRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*WorkflowRun, error)
	HasInProgress(ctx context.Context, workflowID uuid.UUID) (bool, error)
	FindInProgressByWorkflowID(ctx context.Context, workflowID uuid.UUID) (*WorkflowRun, error)
}

type WorkflowRunReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*WorkflowRunView, error)
	FindByWorkflowID(
		ctx context.Context,
		workflowID uuid.UUID,
		query paginate.PaginateQuery,
	) ([]WorkflowRunView, int64, error)
	FindAnalyticsByProject(
		ctx context.Context,
		projectID uuid.UUID,
		filter WorkflowRunAnalyticsFilter,
	) (*WorkflowRunAnalytics, error)
}

type WorkflowRunAnalyticsFilter struct {
	WorkflowID *uuid.UUID
	From       *time.Time
	To         *time.Time
}

type WorkflowRunAnalytics struct {
	TotalRuns         int64
	SuccessRate       float64
	SuccessCount      int64
	FailureRate       float64
	FailureCount      int64
	CancelledCount    int64
	RunningCount      int64
	PendingCount      int64
	AverageDurationMS float64
	MinDurationMS     float64
	MaxDurationMS     float64
	LastRunAt         *time.Time
}

type WorkflowRunView struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	ProjectID    uuid.UUID
	Status            Status
	TriggeredBy       TriggeredBy
	TriggeredByUserID *uuid.UUID
	Context           map[string]any
	StartedAt         *time.Time
	FinishedAt        *time.Time
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
