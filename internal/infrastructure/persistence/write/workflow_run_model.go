package write

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// WorkflowRunModel is the persistence mapping for table workflow_runs.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type WorkflowRunModel struct {
	ID                uuid.UUID    `gorm:"column:id;primaryKey"`
	WorkflowID        uuid.UUID    `gorm:"column:workflow_id"`
	Status            string       `gorm:"column:status"`
	TriggeredBy       string       `gorm:"column:triggered_by"`
	TriggeredByUserID *uuid.UUID   `gorm:"column:triggered_by_user_id"`
	Context           dbtype.JSONB `gorm:"column:context"`
	StartedAt         *time.Time   `gorm:"column:started_at"`
	FinishedAt        *time.Time   `gorm:"column:finished_at"`
	Error             string       `gorm:"column:error"`
	CreatedAt         time.Time    `gorm:"column:created_at"`
	UpdatedAt         time.Time    `gorm:"column:updated_at"`
}

func (WorkflowRunModel) TableName() string {
	return "workflow_runs"
}
