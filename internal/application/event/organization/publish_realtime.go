package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{realtime: realtimePublisher}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	userID, err := uuid.Parse(evt.CreatedByUserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	eventType := realtime.EventType(realtime.EntityOrganization, realtime.ActionCreated)
	if err := h.realtime.PublishToUser(ctx, userID, eventType, evt); err != nil {
		log.Printf("centrifugo publish failed type=%s userId=%s: %v", eventType, evt.CreatedByUserID, err)
		return messaging.Retryable(err)
	}
	log.Printf("centrifugo published type=%s userId=%s", eventType, evt.CreatedByUserID)
	return nil
}
