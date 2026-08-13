package write

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// StepRunModel is the persistence mapping for table step_runs.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type StepRunModel struct {
	ID               uuid.UUID    `gorm:"column:id;primaryKey"`
	WorkflowRunID    uuid.UUID    `gorm:"column:workflow_run_id"`
	StepID           uuid.UUID    `gorm:"column:step_id"`
	WorkflowID       uuid.UUID    `gorm:"column:workflow_id"`
	EndpointID       uuid.UUID    `gorm:"column:endpoint_id"`
	OrganizationID   uuid.UUID    `gorm:"column:organization_id"`
	Name             string       `gorm:"column:name"`
	Description      string       `gorm:"column:description"`
	URL              string       `gorm:"column:url"`
	Method           string       `gorm:"column:method"`
	Headers          dbtype.JSONB `gorm:"column:headers"`
	QueryParams      dbtype.JSONB `gorm:"column:query_params"`
	Body             dbtype.JSONB `gorm:"column:body"`
	Timeout          int          `gorm:"column:timeout_ms"`
	RetryOnFailure   bool         `gorm:"column:retry_on_failure"`
	RetryCount       int          `gorm:"column:retry_count"`
	RetryDelay       int          `gorm:"column:retry_delay_ms"`
	StepIndex        string       `gorm:"column:step_index"`
	ExecutionOrder   int          `gorm:"column:execution_order"`
	TreeIndex        int          `gorm:"column:tree_index"`
	PositionX        float64      `gorm:"column:position_x"`
	PositionY        float64      `gorm:"column:position_y"`
	Status           string       `gorm:"column:status"`
	Attempt          int          `gorm:"column:attempt"`
	ResponseSnapshot dbtype.JSONB `gorm:"column:response_snapshot"`
	StartedAt        *time.Time   `gorm:"column:started_at"`
	FinishedAt       *time.Time   `gorm:"column:finished_at"`
	Error            string       `gorm:"column:error"`
	CreatedAt        time.Time    `gorm:"column:created_at"`
	UpdatedAt        time.Time    `gorm:"column:updated_at"`
}

func (StepRunModel) TableName() string {
	return "step_runs"
}
