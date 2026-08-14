package variable

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type VariableWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, variable *Variable) error
	Update(ctx context.Context, variable *Variable) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByStepID(ctx context.Context, stepID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Variable, error)
}

type VariableReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*VariableView, error)
	FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]VariableView, error)
	FindByStepID(ctx context.Context, stepID uuid.UUID) ([]VariableView, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]VariableView, error)
}

type VariableView struct {
	ID           uuid.UUID
	Name         string
	Key          string
	Description  string
	Path         string
	StepID       uuid.UUID
	WorkflowID   uuid.UUID
	IsSecret     bool
	DefaultValue json.RawMessage
	LastValue    json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
