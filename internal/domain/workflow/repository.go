package workflow

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type WorkflowWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, workflow *Workflow) error
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workflow, error)
}

type WorkflowReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*WorkflowView, error)
	FindByOrganizationID(
		ctx context.Context,
		organizationID uuid.UUID,
		query paginate.PaginateQuery,
	) ([]WorkflowView, int64, error)
}

type WorkflowView struct {
	ID                      uuid.UUID
	Name                    string
	Description             string
	Status                  Status
	OrganizationID          uuid.UUID
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
