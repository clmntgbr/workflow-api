package subscription

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
	userRepo  domainuser.UserReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	userRepo domainuser.UserReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher, nil),
		userRepo:  userRepo,
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

	return h.publisher.ToUser(ctx, realtime.EntitySubscription, realtime.ActionUpdated, user.ID.String(), evt)
}
