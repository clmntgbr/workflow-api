package subscription

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainsubscription "go-api/internal/domain/subscription"
	"go-api/internal/domain/port"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
	userRepo domainuser.UserReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	userRepo domainuser.UserReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime: realtimePublisher,
		userRepo: userRepo,
	}
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainsubscription.SubscriptionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	subscriptionID, err := uuid.Parse(evt.SubscriptionID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	user, err := h.userRepo.FindBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if user == nil {
		log.Printf(
			"subscription realtime skipped: no user linked subscriptionId=%s eventId=%s",
			evt.SubscriptionID,
			evt.ID,
		)
		return nil
	}

	eventType := realtime.EventType(realtime.EntitySubscription, realtime.ActionUpdated)
	if err := h.realtime.PublishToUser(ctx, user.ID, eventType, evt); err != nil {
		log.Printf(
			"centrifugo publish failed type=%s subscriptionId=%s userId=%s: %v",
			eventType,
			evt.SubscriptionID,
			user.ID.String(),
			err,
		)
		return messaging.Retryable(err)
	}

	log.Printf(
		"centrifugo published type=%s subscriptionId=%s userId=%s",
		eventType,
		evt.SubscriptionID,
		user.ID.String(),
	)
	return nil
}
