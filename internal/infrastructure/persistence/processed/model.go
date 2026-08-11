package processed

import (
	"time"

	"github.com/google/uuid"
)

type ProcessedEvent struct {
	EventID     uuid.UUID `gorm:"column:event_id;primaryKey"`
	HandlerName string    `gorm:"column:handler_name;primaryKey"`
	ProcessedAt time.Time `gorm:"column:processed_at"`
}

func (ProcessedEvent) TableName() string {
	return "processed_events"
}
