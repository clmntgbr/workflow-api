package subscription

import (
	"context"
	"errors"

	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type SubscriptionDeletedCommand struct {
	StripeSubscriptionID string
}

type SubscriptionDeletedHandler struct {
	planRepo         plan.PlanWriteRepository
	subscriptionRepo domainsubscription.SubscriptionWriteRepository
	outbox           port.OutboxRepository
}

func NewSubscriptionDeletedHandler(
	planRepo plan.PlanWriteRepository,
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
) *SubscriptionDeletedHandler {
	return &SubscriptionDeletedHandler{
		planRepo:         planRepo,
		subscriptionRepo: subscriptionRepo,
		outbox:           outbox,
	}
}

func (h *SubscriptionDeletedHandler) Handle(ctx context.Context, cmd SubscriptionDeletedCommand) error {
	if cmd.StripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	subscriptionEntity, err := h.subscriptionRepo.GetByStripeSubscriptionID(ctx, cmd.StripeSubscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscriptionEntity == nil {
		return nil
	}

	freePlan, err := h.planRepo.GetBySlug(ctx, plan.FreePlanSlug)
	if err != nil {
		return errors.New("failed to get free plan")
	}
	if freePlan == nil {
		return errors.New("free plan not found")
	}

	subscriptionEntity.DowngradeToFree(freePlan.ID)

	return h.subscriptionRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
			return errors.New("failed to update subscription")
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
}
