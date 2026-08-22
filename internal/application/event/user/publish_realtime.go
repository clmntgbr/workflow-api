package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{realtime: realtimePublisher}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainuser.UserCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publish(ctx, evt.UserID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainuser.UserUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publish(ctx, evt.UserID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) OnDeleted(ctx context.Context, payload []byte) error {
	var evt domainuser.UserDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publish(ctx, evt.UserID, realtime.ActionDeleted, evt)
}

func (h *PublishRealtimeHandler) OnActiveProjectChanged(ctx context.Context, payload []byte) error {
	var evt domainuser.UserActiveProjectChanged
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publish(ctx, evt.UserID, realtime.ActionActiveProjectChanged, evt)
}

func (h *PublishRealtimeHandler) publish(ctx context.Context, userIDRaw, action string, payload any) error {
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	eventType := realtime.EventType(realtime.EntityUser, action)
	if err := h.realtime.PublishToUser(ctx, userID, eventType, payload); err != nil {
		log.Printf("centrifugo publish failed type=%s userId=%s: %v", eventType, userIDRaw, err)
		return messaging.Retryable(err)
	}
	log.Printf("centrifugo published type=%s userId=%s", eventType, userIDRaw)
	return nil
}
