package subscription

import (
	"context"
	"errors"
	"log"

	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type InvoicePaymentFailedCommand struct {
	StripeSubscriptionID string
	StripeCustomerID     string
}

type InvoicePaymentFailedHandler struct {
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewInvoicePaymentFailedHandler(
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *InvoicePaymentFailedHandler {
	return &InvoicePaymentFailedHandler{
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *InvoicePaymentFailedHandler) Handle(ctx context.Context, cmd InvoicePaymentFailedCommand) error {
	if cmd.StripeSubscriptionID == "" {
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
		log.Printf(
			"invoice payment failed: subscription not linked yet stripeSubscriptionID=%s stripeCustomerID=%s",
			cmd.StripeSubscriptionID,
			cmd.StripeCustomerID,
		)
		return ErrStripeSubscriptionNotLinked
	}

	stripeCustomerID := subscriptionEntity.StripeCustomerID
	if cmd.StripeCustomerID != "" {
		stripeCustomerID = cmd.StripeCustomerID
	}

	subscriptionEntity.ApplyUpdate(
		subscriptionEntity.PlanID,
		domainsubscription.StatusPastDue,
		stripeCustomerID,
		cmd.StripeSubscriptionID,
		subscriptionEntity.StartDate,
		subscriptionEntity.EndDate,
		subscriptionEntity.CancelAtPeriodEnd,
		subscriptionEntity.QuotaPeriodStart,
	)

	return h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
}
