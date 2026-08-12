package outbox

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/persistence/dbtype"
	"go-api/internal/infrastructure/persistence/write"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) port.OutboxRepository {
	return &repository{db: db}
}

func (r *repository) StoreEvents(ctx context.Context, events []event.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	rows := make([]OutboxEvent, 0, len(events))
	for _, e := range events {
		payload, err := json.Marshal(e)
		if err != nil {
			return err
		}
		eventID, err := uuid.Parse(e.EventID())
		if err != nil {
			return err
		}
		rows = append(rows, OutboxEvent{
			ID:          eventID,
			AggregateID: e.AggregateID(),
			EventType:   e.EventType(),
			Payload:     dbtype.JSONB(payload),
			CreatedAt:   e.OccurredAt(),
		})
	}

	return write.DBWithContext(ctx, r.db).Create(&rows).Error
}

func (r *repository) FetchUnpublished(ctx context.Context, limit int) ([]port.OutboxMessage, error) {
	var rows []OutboxEvent
	err := r.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	messages := make([]port.OutboxMessage, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, port.OutboxMessage{
			ID:          row.ID,
			AggregateID: row.AggregateID,
			EventType:   row.EventType,
			Payload:     []byte(row.Payload),
			CreatedAt:   row.CreatedAt,
			PublishedAt: row.PublishedAt,
		})
	}
	return messages, nil
}

func (r *repository) MarkPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&OutboxEvent{}).
		Where("id IN ?", ids).
		Update("published_at", now).Error
}
