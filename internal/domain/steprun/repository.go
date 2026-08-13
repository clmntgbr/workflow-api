package steprun

import (
	"context"
	"time"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type StepRunWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, run *StepRun) error
	Update(ctx context.Context, run *StepRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*StepRun, error)
	FindByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]*StepRun, error)
}

type StepRunReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*StepRunView, error)
	FindByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]StepRunView, error)
}

type StepRunView struct {
	ID               uuid.UUID
	WorkflowRunID    uuid.UUID
	StepID           uuid.UUID
	WorkflowID       uuid.UUID
	EndpointID       uuid.UUID
	OrganizationID   uuid.UUID
	Name             string
	Description      string
	URL              string
	Method           string
	Headers          map[string]string
	Query            map[string]string
	Body             map[string]any
	Timeout          int
	RetryOnFailure   bool
	RetryCount       int
	RetryDelay       int
	Index            string
	ExecutionOrder   int
	TreeIndex        int
	Position         domainstep.Position
	Status           Status
	Attempt          int
	ResponseSnapshot *ResponseSnapshot
	StartedAt        *time.Time
	FinishedAt       *time.Time
	Error            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
