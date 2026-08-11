package port

import (
	"context"

	"github.com/google/uuid"
)

type RealtimePublisher interface {
	PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, payload any) error
}
