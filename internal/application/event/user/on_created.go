package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainuser "go-api/internal/domain/user"
)

type UserCreatedHandler struct{}

func NewUserCreatedHandler() *UserCreatedHandler {
	return &UserCreatedHandler{}
}

func (h *UserCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s userId=%s clerkId=%s email=%s",
		domainuser.EventTypeUserCreated,
		evt.ID,
		evt.UserID,
		evt.ClerkID,
		evt.Email,
	)
	return nil
}
