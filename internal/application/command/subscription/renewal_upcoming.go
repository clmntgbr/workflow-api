package subscription

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type SubscriptionRenewalUpcomingCommand struct {
	StripeEventID        string
	StripeSubscriptionID string
	StripeCustomerID     string
	AmountDue            int64
	Currency             string
	PeriodEnd            time.Time
}

type SubscriptionRenewalUpcomingHandler struct {
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewSubscriptionRenewalUpcomingHandler(
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *SubscriptionRenewalUpcomingHandler {
	return &SubscriptionRenewalUpcomingHandler{
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *SubscriptionRenewalUpcomingHandler) Handle(ctx context.Context, cmd SubscriptionRenewalUpcomingCommand) error {
	if cmd.StripeSubscriptionID == "" && cmd.StripeCustomerID == "" {
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

	subscriptionEntity.RecordRenewalUpcoming(
		event.DeterministicID("stripe", cmd.StripeEventID, domainsubscription.EventTypeSubscriptionRenewalUpcoming),
		cmd.AmountDue,
		cmd.Currency,
		cmd.PeriodEnd,
	)

	return h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
}
