package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainuser "go-api/internal/domain/user"
)

type UserActiveProjectChangedHandler struct{}

func NewUserActiveProjectChangedHandler() *UserActiveProjectChangedHandler {
	return &UserActiveProjectChangedHandler{}
}

func (h *UserActiveProjectChangedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserActiveProjectChanged
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s userId=%s projectId=%s",
		domainuser.EventTypeUserActiveProjectChanged,
		evt.ID,
		evt.UserID,
		evt.ProjectID,
	)
	return nil
}
