package subscription

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type CheckoutCompletedCommand struct {
	UserID               uuid.UUID
	StripeCustomerID     string
	StripeSubscriptionID string
}

type CheckoutCompletedHandler struct {
	userRepo            domainuser.UserWriteRepository
	planRepo            plan.PlanWriteRepository
	subscriptionRepo    domainsubscription.SubscriptionWriteRepository
	outbox              port.OutboxRepository
	subscriptionGateway port.SubscriptionGateway
}

func NewCheckoutCompletedHandler(
	userRepo domainuser.UserWriteRepository,
	planRepo plan.PlanWriteRepository,
	subscriptionRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
	subscriptionGateway port.SubscriptionGateway,
) *CheckoutCompletedHandler {
	return &CheckoutCompletedHandler{
		userRepo:            userRepo,
		planRepo:            planRepo,
		subscriptionRepo:    subscriptionRepo,
		outbox:              outbox,
		subscriptionGateway: subscriptionGateway,
	}
}

func (h *CheckoutCompletedHandler) Handle(ctx context.Context, cmd CheckoutCompletedCommand) error {
	if cmd.StripeSubscriptionID == "" {
		return errors.New("stripe subscription id is required")
	}

	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return errors.New("failed to get user")
	}
	if user == nil {
		return errors.New("user not found")
	}

	subData, err := h.subscriptionGateway.Retrieve(ctx, cmd.StripeSubscriptionID)
	if err != nil {
		return err
	}

	targetPlan, err := h.planRepo.GetByStripePriceID(ctx, subData.PriceID)
	if err != nil {
		return errors.New("failed to get plan")
	}
	if targetPlan == nil {
		return errors.New("plan not found for stripe price id")
	}

	customerID := cmd.StripeCustomerID
	if customerID == "" {
		customerID = subData.CustomerID
	}

	status := domainsubscription.MapBillingStatus(subData.Status)

	var subscriptionEntity *domainsubscription.Subscription
	if user.SubscriptionID != nil {
		subscriptionEntity, err = h.subscriptionRepo.GetByID(ctx, *user.SubscriptionID)
		if err != nil {
			return errors.New("failed to get subscription")
		}
	}

	wasFree := subscriptionEntity == nil
	if subscriptionEntity != nil {
		currentPlan, err := h.planRepo.GetByID(ctx, subscriptionEntity.PlanID)
		if err != nil {
			return errors.New("failed to get current plan")
		}
		wasFree = currentPlan == nil || currentPlan.Slug == plan.FreePlanSlug || subscriptionEntity.StripeSubscriptionID == ""
	}

	startDate := time.Now().UTC()
	endDate := startDate
	if !subData.CurrentPeriodStart.IsZero() {
		startDate = subData.CurrentPeriodStart
	}
	if !subData.CurrentPeriodEnd.IsZero() {
		endDate = subData.CurrentPeriodEnd
	}

	quotaPeriodStart := time.Time{}
	if wasFree {
		if !subData.CurrentPeriodStart.IsZero() {
			quotaPeriodStart = subData.CurrentPeriodStart
		} else {
			quotaPeriodStart = time.Now().UTC()
		}
	}

	return h.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if subscriptionEntity == nil {
			subscriptionEntity = domainsubscription.NewSubscription(targetPlan.ID, status, startDate, endDate)
			subscriptionEntity.ApplyUpdate(
				targetPlan.ID,
				status,
				customerID,
				subData.ID,
				startDate,
				endDate,
				subData.CancelAtPeriodEnd,
				quotaPeriodStart,
			)
			if err := h.subscriptionRepo.Save(txCtx, subscriptionEntity); err != nil {
				return errors.New("failed to create subscription")
			}
		} else {
			if quotaPeriodStart.IsZero() {
				quotaPeriodStart = subscriptionEntity.QuotaPeriodStart
			}
			subscriptionEntity.ApplyUpdate(
				targetPlan.ID,
				status,
				customerID,
				subData.ID,
				startDate,
				endDate,
				subData.CancelAtPeriodEnd,
				quotaPeriodStart,
			)
			if err := h.subscriptionRepo.Update(txCtx, subscriptionEntity); err != nil {
				return errors.New("failed to update subscription")
			}
		}

		events := subscriptionEntity.PullEvents()

		if user.SubscriptionID == nil || *user.SubscriptionID != subscriptionEntity.ID {
			user.AssignSubscription(subscriptionEntity.ID)
			if err := h.userRepo.Update(txCtx, user); err != nil {
				return errors.New("failed to link subscription to user")
			}
			events = append(events, user.PullEvents()...)
		}

		return h.outbox.StoreEvents(txCtx, events)
	})
}
