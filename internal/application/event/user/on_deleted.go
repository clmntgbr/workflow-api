package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainuser "go-api/internal/domain/user"
)

type UserDeletedHandler struct{}

func NewUserDeletedHandler() *UserDeletedHandler {
	return &UserDeletedHandler{}
}

func (h *UserDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s userId=%s clerkId=%s",
		domainuser.EventTypeUserDeleted,
		evt.ID,
		evt.UserID,
		evt.ClerkID,
	)
	return nil
}
