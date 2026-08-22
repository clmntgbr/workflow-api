package subscription

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type SubscriptionUpdatedCommand struct {
	StripeSubscriptionID string
	StripeCustomerID     string
	StripePriceID        string
	Status               string
	CancelAtPeriodEnd    bool
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
}

type SubscriptionUpdatedHandler struct {
	planRepo         plan.PlanWriteRepository
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewSubscriptionUpdatedHandler(
	planRepo plan.PlanWriteRepository,
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *SubscriptionUpdatedHandler {
	return &SubscriptionUpdatedHandler{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *SubscriptionUpdatedHandler) Handle(ctx context.Context, cmd SubscriptionUpdatedCommand) error {
	if cmd.StripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	subscriptionEntity, err := h.subscriptionRepo.GetByStripeSubscriptionID(ctx, cmd.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}

	if subscriptionEntity == nil && cmd.StripeCustomerID != "" {
		subscriptionEntity, err = h.subscriptionRepo.GetByStripeCustomerID(ctx, cmd.StripeCustomerID)
		if err != nil {
			return errors.New("failed to get subscription by customer")
		}
	}

	if subscriptionEntity == nil {
		return nil
	}

	planID := subscriptionEntity.PlanID
	if cmd.StripePriceID != "" {
		targetPlan, err := h.planRepo.GetByStripePriceID(ctx, cmd.StripePriceID)
		if err != nil {
			return errors.New("failed to get plan")
		}
		if targetPlan != nil {
			planID = targetPlan.ID
		}
	}

	stripeCustomerID := subscriptionEntity.StripeCustomerID
	if cmd.StripeCustomerID != "" {
		stripeCustomerID = cmd.StripeCustomerID
	}

	startDate := subscriptionEntity.StartDate
	if !cmd.CurrentPeriodStart.IsZero() {
		startDate = cmd.CurrentPeriodStart
	}
	endDate := subscriptionEntity.EndDate
	if !cmd.CurrentPeriodEnd.IsZero() {
		endDate = cmd.CurrentPeriodEnd
	}

	subscriptionEntity.ApplyUpdate(
		planID,
		domainsubscription.MapBillingStatus(cmd.Status),
		stripeCustomerID,
		cmd.StripeSubscriptionID,
		startDate,
		endDate,
		cmd.CancelAtPeriodEnd,
		subscriptionEntity.QuotaPeriodStart,
	)

	return h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
}
