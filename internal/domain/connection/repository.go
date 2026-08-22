package connection

import (
	"context"

	"github.com/google/uuid"
)

type ConnectionWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, conn *Connection) error
	GetByID(ctx context.Context, id uuid.UUID) (*Connection, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ConnectionReadRepository interface {
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]ConnectionView, error)
}

type ConnectionView struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
	SourceStepID   uuid.UUID
	TargetStepID   uuid.UUID
}
