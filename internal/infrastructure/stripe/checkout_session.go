package stripe

import (
	"context"
	"fmt"
	"strconv"

	"go-api/internal/infrastructure/config"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type CheckoutSessionGateway struct {
	redirectSuccessURL string
	redirectCancelURL  string
	secretKey          string
}

func NewCheckoutSessionGateway(cfg *config.Config) *CheckoutSessionGateway {
	return &CheckoutSessionGateway{
		redirectSuccessURL: cfg.RedirectSuccessURL,
		redirectCancelURL:  cfg.RedirectCancelURL,
		secretKey:          cfg.StripeSecretKey,
	}
}

func (g *CheckoutSessionGateway) Create(
	ctx context.Context,
	planID string,
	planName string,
	planPrice float64,
	currency string,
	billingInterval string,
	stripePriceID string,
	userID string,
	userFirstName string,
	userLastName string,
	email string,
	stripeCustomerID string,
) (string, error) {
	if g.secretKey == "" {
		return "", fmt.Errorf("stripe secret key is not configured")
	}
	if stripePriceID == "" {
		return "", fmt.Errorf("plan has no stripe price id")
	}

	stripe.Key = g.secretKey

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(userID),
		SuccessURL:        stripe.String(g.redirectSuccessURL),
		CancelURL:         stripe.String(g.redirectCancelURL),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(stripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"plan_id":       planID,
			"plan_name":     planName,
			"plan_price":    strconv.FormatFloat(planPrice, 'f', -1, 64),
			"currency":      currency,
			"plan_interval": billingInterval,
			"user_id":       userID,
			"user_email":    email,
			"user_name":     userFirstName + " " + userLastName,
		},
	}
	params.Context = ctx

	if stripeCustomerID != "" {
		params.Customer = stripe.String(stripeCustomerID)
	} else if email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create checkout session: %w", err)
	}
	if s.URL == "" {
		return "", fmt.Errorf("checkout session URL is required")
	}

	return s.URL, nil
}
