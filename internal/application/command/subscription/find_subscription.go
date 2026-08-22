package subscription

import (
	"context"
	"errors"

	domainsubscription "go-api/internal/domain/subscription"
)

func findSubscriptionByStripeIDs(
	ctx context.Context,
	repo domainsubscription.SubscriptionWriteRepository,
	stripeSubscriptionID string,
	stripeCustomerID string,
) (*domainsubscription.Subscription, error) {
	var (
		subscriptionEntity *domainsubscription.Subscription
		err                error
	)

	if stripeSubscriptionID != "" {
		subscriptionEntity, err = repo.GetByStripeSubscriptionID(ctx, stripeSubscriptionID)
		if err != nil {
			return nil, errors.New("failed to get subscription")
		}
	}

	if subscriptionEntity == nil && stripeCustomerID != "" {
		subscriptionEntity, err = repo.GetByStripeCustomerID(ctx, stripeCustomerID)
		if err != nil {
			return nil, errors.New("failed to get subscription by customer")
		}
	}

	return subscriptionEntity, nil
}
