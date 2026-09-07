package runexport

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, job *RunExport) error
	Update(ctx context.Context, job *RunExport) error
	GetByID(ctx context.Context, id uuid.UUID) (*RunExport, error)
}

type ReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*View, error)
}

type View struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	ProjectID         uuid.UUID
	RequestedByUserID uuid.UUID
	DateRangeFrom     time.Time
	DateRangeTo       time.Time
	Status            Status
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
