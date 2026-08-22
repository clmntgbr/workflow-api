package subscription

import (
	"context"
	"errors"
	"log"

	identitycmd "go-api/internal/application/command/identity"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type CreateSubscriptionCommand struct {
	UserID        uuid.UUID
	PlanID        uuid.UUID
	ProrationDate *int64
}

type CreateSubscriptionResult struct {
	URL     string
	Updated bool
}

type CreateSubscriptionHandler struct {
	userRepo               domainuser.UserReadRepository
	planRepo               plan.PlanReadRepository
	subscriptionReadRepo   domainsubscription.SubscriptionReadRepository
	subscriptionWriteRepo  domainsubscription.SubscriptionWriteRepository
	outbox                 port.OutboxRepository
	fetchUser              *identitycmd.FetchUserHandler
	checkoutSessionGateway port.CheckoutSessionGateway
	subscriptionGateway    port.SubscriptionGateway
}

func NewCreateSubscriptionHandler(
	userRepo domainuser.UserReadRepository,
	planRepo plan.PlanReadRepository,
	subscriptionReadRepo domainsubscription.SubscriptionReadRepository,
	subscriptionWriteRepo domainsubscription.SubscriptionWriteRepository,
	outbox port.OutboxRepository,
	fetchUser *identitycmd.FetchUserHandler,
	checkoutSessionGateway port.CheckoutSessionGateway,
	subscriptionGateway port.SubscriptionGateway,
) *CreateSubscriptionHandler {
	return &CreateSubscriptionHandler{
		userRepo:               userRepo,
		planRepo:               planRepo,
		subscriptionReadRepo:   subscriptionReadRepo,
		subscriptionWriteRepo:  subscriptionWriteRepo,
		outbox:                 outbox,
		fetchUser:              fetchUser,
		checkoutSessionGateway: checkoutSessionGateway,
		subscriptionGateway:    subscriptionGateway,
	}
}

func (h *CreateSubscriptionHandler) Handle(
	ctx context.Context,
	cmd CreateSubscriptionCommand,
) (*CreateSubscriptionResult, error) {
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	targetPlan, err := h.planRepo.FindByID(ctx, cmd.PlanID)
	if err != nil {
		return nil, errors.New("failed to get plan")
	}
	if targetPlan == nil {
		return nil, querysubscription.ErrPlanNotFound
	}
	if !targetPlan.IsActive {
		return nil, querysubscription.ErrPlanInactive
	}
	if targetPlan.Slug == plan.FreePlanSlug {
		return nil, querysubscription.ErrFreePlanCheckout
	}
	if targetPlan.StripePriceID == "" {
		return nil, querysubscription.ErrMissingStripePrice
	}

	var current *domainsubscription.SubscriptionView
	if user.SubscriptionID != nil {
		current, err = h.subscriptionReadRepo.FindByID(ctx, *user.SubscriptionID)
		if err != nil {
			return nil, errors.New("failed to get subscription")
		}
	}

	if current != nil && current.PlanID == targetPlan.ID && current.StripeSubscriptionID != "" {
		return nil, querysubscription.ErrAlreadyOnPlan
	}

	if querysubscription.CanUpdateStripeSubscription(current) {
		if err := h.updateExistingSubscription(ctx, current.ID, targetPlan, cmd.ProrationDate); err != nil {
			return nil, err
		}
		return &CreateSubscriptionResult{Updated: true}, nil
	}

	clerkUser, err := h.fetchUser.Handle(ctx, user.ClerkID)
	if err != nil {
		return nil, errors.New("failed to get user email")
	}

	stripeCustomerID := ""
	if current != nil {
		stripeCustomerID = current.StripeCustomerID
	}

	url, err := h.checkoutSessionGateway.Create(
		ctx,
		targetPlan.ID.String(),
		targetPlan.Name,
		targetPlan.Price,
		string(targetPlan.Currency),
		string(targetPlan.BillingInterval),
		targetPlan.StripePriceID,
		user.ID.String(),
		user.FirstName,
		user.LastName,
		clerkUser.Email,
		stripeCustomerID,
	)
	if err != nil {
		return nil, err
	}

	return &CreateSubscriptionResult{URL: url}, nil
}

func (h *CreateSubscriptionHandler) updateExistingSubscription(
	ctx context.Context,
	subscriptionID uuid.UUID,
	targetPlan *plan.PlanView,
	prorationDate *int64,
) error {
	subscriptionEntity, err := h.subscriptionWriteRepo.GetByID(ctx, subscriptionID)
	if err != nil {
		return errors.New("failed to get subscription")
	}
	if subscriptionEntity == nil {
		return querysubscription.ErrSubscriptionNotFound
	}
	if subscriptionEntity.StripeSubscriptionID == "" {
		return querysubscription.ErrMissingStripeSub
	}

	stripeSub, err := h.subscriptionGateway.Retrieve(ctx, subscriptionEntity.StripeSubscriptionID)
	if err != nil {
		return err
	}
	if stripeSub.ItemID == "" {
		return querysubscription.ErrMissingStripeSub
	}
	if stripeSub.PriceID == targetPlan.StripePriceID {
		return querysubscription.ErrAlreadyOnPlan
	}

	updated, err := h.subscriptionGateway.UpdatePrice(
		ctx,
		subscriptionEntity.StripeSubscriptionID,
		stripeSub.ItemID,
		targetPlan.StripePriceID,
		prorationDate,
	)
	if err != nil {
		return err
	}

	stripeCustomerID := subscriptionEntity.StripeCustomerID
	if updated.CustomerID != "" {
		stripeCustomerID = updated.CustomerID
	}

	startDate := subscriptionEntity.StartDate
	if !updated.CurrentPeriodStart.IsZero() {
		startDate = updated.CurrentPeriodStart
	}
	endDate := subscriptionEntity.EndDate
	if !updated.CurrentPeriodEnd.IsZero() {
		endDate = updated.CurrentPeriodEnd
	}

	subscriptionEntity.ApplyUpdate(
		targetPlan.ID,
		domainsubscription.MapBillingStatus(updated.Status),
		stripeCustomerID,
		subscriptionEntity.StripeSubscriptionID,
		startDate,
		endDate,
		updated.CancelAtPeriodEnd,
		subscriptionEntity.QuotaPeriodStart,
	)

	err = h.subscriptionWriteRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.subscriptionWriteRepo.Update(txCtx, subscriptionEntity); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, subscriptionEntity.PullEvents())
	})
	if err != nil {
		return errors.New("failed to update subscription")
	}

	log.Printf(
		"subscription plan changed locally id=%s plan=%s stripeSub=%s",
		subscriptionEntity.ID,
		targetPlan.Slug,
		subscriptionEntity.StripeSubscriptionID,
	)

	return nil
}
