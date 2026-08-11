package user

import (
	"context"
	"encoding/json"
	"fmt"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"
)

type NotifyUserOnCreatedHandler struct {
	notifier port.NotificationSender
}

func NewNotifyUserOnCreatedHandler(notifier port.NotificationSender) *NotifyUserOnCreatedHandler {
	return &NotifyUserOnCreatedHandler{notifier: notifier}
}

func (h *NotifyUserOnCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	msg := fmt.Sprintf("User created: %s %s (%s)", evt.FirstName, evt.LastName, evt.Email)
	if err := h.notifier.Send(ctx, evt.UserID, msg); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}
