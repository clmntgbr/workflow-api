package step

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type StepWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, step *Step) error
	Update(ctx context.Context, step *Step) error
	GetByID(ctx context.Context, id uuid.UUID) (*Step, error)
}

type StepReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*StepView, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]StepView, error)
}

type StepView struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       Position
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
