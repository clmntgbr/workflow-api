package port

import (
	"context"
	"encoding/json"
)

type EventEnvelope struct {
	EventID     string          `json:"eventId"`
	Type        string          `json:"type"`
	AggregateID string          `json:"aggregateId"`
	OccurredAt  string          `json:"occurredAt"`
	Payload     json.RawMessage `json:"payload"`
	Priority    uint8           `json:"-"`
}

type EventPublisher interface {
	Publish(ctx context.Context, envelope EventEnvelope) error
}
