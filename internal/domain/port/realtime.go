package port

import (
	"context"

	"github.com/google/uuid"
)

type RealtimePublisher interface {
	PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, payload any) error
}

// RealtimeConnection carries the credentials a client needs to open a realtime
// connection. It is transport- and vendor-neutral: the HTTP shape lives in the
// presenter, the provider specifics in the infrastructure adapter.
type RealtimeConnection struct {
	Token   string
	Channel string
	WSURL   string
}
