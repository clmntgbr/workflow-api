package subscription

import (
	"context"
	"errors"
	"log"
	"time"

	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type InvoicePaymentSucceededCommand struct {
	StripeSubscriptionID string
	StripeCustomerID     string
	BillingReason        string
	PeriodStart          time.Time
	PeriodEnd            time.Time
}

type InvoicePaymentSucceededHandler struct {
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewInvoicePaymentSucceededHandler(
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *InvoicePaymentSucceededHandler {
	return &InvoicePaymentSucceededHandler{
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *InvoicePaymentSucceededHandler) Handle(ctx context.Context, cmd InvoicePaymentSucceededCommand) error {
	log.Printf(
		"invoice payment succeeded: start stripeSubscriptionID=%s stripeCustomerID=%s billingReason=%s",
		cmd.StripeSubscriptionID,
		cmd.StripeCustomerID,
		cmd.BillingReason,
	)

	if cmd.StripeSubscriptionID == "" {
		log.Printf("invoice payment succeeded: skip, missing stripeSubscriptionID")
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
			"invoice payment succeeded: subscription not linked yet stripeSubscriptionID=%s stripeCustomerID=%s",
			cmd.StripeSubscriptionID,
			cmd.StripeCustomerID,
		)
		return ErrStripeSubscriptionNotLinked
	}

	stripeCustomerID := subscriptionEntity.StripeCustomerID
	if cmd.StripeCustomerID != "" {
		stripeCustomerID = cmd.StripeCustomerID
	}

	startDate := subscriptionEntity.StartDate
	if !cmd.PeriodStart.IsZero() {
		startDate = cmd.PeriodStart
	}
	endDate := subscriptionEntity.EndDate
	if !cmd.PeriodEnd.IsZero() {
		endDate = cmd.PeriodEnd
	}

	subscriptionEntity.ApplyUpdate(
		subscriptionEntity.PlanID,
		domainsubscription.StatusActive,
		stripeCustomerID,
		cmd.StripeSubscriptionID,
		startDate,
		endDate,
		false,
		subscriptionEntity.QuotaPeriodStart,
	)

	err = h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
	if err != nil {
		log.Printf("invoice payment succeeded: update failed subscriptionID=%s: %v", subscriptionEntity.ID, err)
		return err
	}

	log.Printf("invoice payment succeeded: updated subscriptionID=%s status=active", subscriptionEntity.ID)
	return nil
}
