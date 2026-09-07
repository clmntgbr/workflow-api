package subscription

import (
	"context"
	"errors"

	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type PaymentMethodExpiringCommand struct {
	StripeEventID        string
	StripeCustomerID     string
	StripeSubscriptionID string
	Brand                string
	Last4                string
	ExpMonth             int64
	ExpYear              int64
}

type PaymentMethodExpiringHandler struct {
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewPaymentMethodExpiringHandler(
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *PaymentMethodExpiringHandler {
	return &PaymentMethodExpiringHandler{
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *PaymentMethodExpiringHandler) Handle(ctx context.Context, cmd PaymentMethodExpiringCommand) error {
	if cmd.StripeCustomerID == "" && cmd.StripeSubscriptionID == "" {
		return nil
	}

	subscriptionEntity, err := findSubscriptionByStripeIDs(
		ctx,
		h.subscriptionRepo,
		cmd.StripeSubscriptionID,
		cmd.StripeCustomerID,
	)
	if err != nil {
		return err
	}
	if subscriptionEntity == nil {
		return ErrStripeSubscriptionNotLinked
	}

	subscriptionEntity.RecordPaymentMethodExpiring(
		event.DeterministicID("stripe", cmd.StripeEventID, domainsubscription.EventTypeSubscriptionPaymentMethodExpiring),
		cmd.Brand,
		cmd.Last4,
		cmd.ExpMonth,
		cmd.ExpYear,
	)

	return h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
}
