package stripe

import (
	"context"
	"fmt"

	"go-api/internal/infrastructure/config"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/billingportal/session"
)

type BillingPortalGateway struct {
	returnURL string
	secretKey string
}

func NewBillingPortalGateway(cfg *config.Config) *BillingPortalGateway {
	return &BillingPortalGateway{
		returnURL: cfg.RedirectPortalURL,
		secretKey: cfg.StripeSecretKey,
	}
}

func (g *BillingPortalGateway) Create(ctx context.Context, stripeCustomerID string) (string, error) {
	if g.secretKey == "" {
		return "", fmt.Errorf("stripe secret key is not configured")
	}
	if stripeCustomerID == "" {
		return "", fmt.Errorf("stripe customer id is required")
	}
	if g.returnURL == "" {
		return "", fmt.Errorf("stripe portal return url is not configured")
	}

	stripe.Key = g.secretKey

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(stripeCustomerID),
		ReturnURL: stripe.String(g.returnURL),
	}
	params.Context = ctx

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create billing portal session: %w", err)
	}
	if s.URL == "" {
		return "", fmt.Errorf("billing portal session URL is required")
	}

	return s.URL, nil
}
