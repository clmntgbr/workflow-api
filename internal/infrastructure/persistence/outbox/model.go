package outbox

import (
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

// OutboxEvent is the persistence mapping for table outbox_events.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type OutboxEvent struct {
	ID          uuid.UUID    `gorm:"column:id;primaryKey"`
	AggregateID string       `gorm:"column:aggregate_id"`
	EventType   string       `gorm:"column:event_type"`
	Payload     dbtype.JSONB `gorm:"column:payload"`
	CreatedAt   time.Time    `gorm:"column:created_at"`
	PublishedAt *time.Time   `gorm:"column:published_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
