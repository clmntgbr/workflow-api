package subscription

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/plan"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

type CreateBillingPortalCommand struct {
	UserID uuid.UUID
}

type CreateBillingPortalHandler struct {
	userRepo             domainuser.UserReadRepository
	subscriptionRepo     domainsubscription.SubscriptionReadRepository
	billingPortalGateway port.BillingPortalGateway
}

func NewCreateBillingPortalHandler(
	userRepo domainuser.UserReadRepository,
	subscriptionRepo domainsubscription.SubscriptionReadRepository,
	billingPortalGateway port.BillingPortalGateway,
) *CreateBillingPortalHandler {
	return &CreateBillingPortalHandler{
		userRepo:             userRepo,
		subscriptionRepo:     subscriptionRepo,
		billingPortalGateway: billingPortalGateway,
	}
}

func (h *CreateBillingPortalHandler) Handle(ctx context.Context, cmd CreateBillingPortalCommand) (string, error) {
	user, err := h.userRepo.FindByID(ctx, cmd.UserID)
	if err != nil {
		return "", errors.New("failed to get user")
	}
	if user == nil || user.SubscriptionID == nil {
		return "", querysubscription.ErrSubscriptionNotFound
	}

	subscriptionView, err := h.subscriptionRepo.FindByID(ctx, *user.SubscriptionID)
	if err != nil {
		return "", errors.New("failed to get subscription")
	}
	if subscriptionView == nil {
		return "", querysubscription.ErrSubscriptionNotFound
	}
	if subscriptionView.StripeCustomerID == "" {
		if subscriptionView.Plan != nil && subscriptionView.Plan.Slug == plan.FreePlanSlug {
			return "", ErrFreePlanBillingPortal
		}
		return "", ErrMissingStripeCustomer
	}

	url, err := h.billingPortalGateway.Create(ctx, subscriptionView.StripeCustomerID)
	if err != nil {
		return "", err
	}

	return url, nil
}
