package assertion

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AssertionWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, assertion *Assertion) error
	Update(ctx context.Context, assertion *Assertion) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByStepID(ctx context.Context, stepID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Assertion, error)
}

type AssertionReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*AssertionView, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]AssertionView, error)
	FindByStepID(ctx context.Context, stepID uuid.UUID) ([]AssertionView, error)
}

type AssertionView struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Source        AssertionSource
	Path          string
	Operator      AssertionOperator
	ExpectedValue string
	StepID        uuid.UUID
	WorkflowID    uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
