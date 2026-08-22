package subscription

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainsubscription "go-api/internal/domain/subscription"
)

type SubscriptionCreatedHandler struct{}

func NewSubscriptionCreatedHandler() *SubscriptionCreatedHandler {
	return &SubscriptionCreatedHandler{}
}

func (h *SubscriptionCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainsubscription.SubscriptionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s subscriptionId=%s planId=%s status=%s stripeSub=%s",
		domainsubscription.EventTypeSubscriptionCreated,
		evt.ID,
		evt.SubscriptionID,
		evt.PlanID,
		evt.Status,
		evt.StripeSubscriptionID,
	)
	return nil
}
