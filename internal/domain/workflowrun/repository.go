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
}

type WorkflowRunReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*WorkflowRunView, error)
	FindByWorkflowID(
		ctx context.Context,
		workflowID uuid.UUID,
		query paginate.PaginateQuery,
	) ([]WorkflowRunView, int64, error)
}

// WorkflowRunView is the read model. OrganizationID comes from workflows.organization_id
// (the run table does not store a tenant column).
type WorkflowRunView struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	OrganizationID    uuid.UUID
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
