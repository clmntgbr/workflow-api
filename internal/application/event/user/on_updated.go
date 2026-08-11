package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainuser "go-api/internal/domain/user"
)

type UserUpdatedHandler struct{}

func NewUserUpdatedHandler() *UserUpdatedHandler {
	return &UserUpdatedHandler{}
}

func (h *UserUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s userId=%s clerkId=%s banned=%v",
		domainuser.EventTypeUserUpdated,
		evt.ID,
		evt.UserID,
		evt.ClerkID,
		evt.Banned,
	)
	return nil
}
