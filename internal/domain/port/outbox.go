package port

import (
	"context"
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type OutboxMessage struct {
	ID          uuid.UUID
	AggregateID string
	EventType   string
	Payload     []byte
	CreatedAt   time.Time
	PublishedAt *time.Time
}

type OutboxRepository interface {
	StoreEvents(ctx context.Context, events []event.DomainEvent) error
	FetchUnpublished(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID) error
}
