package stripe

import (
	"context"
	"fmt"
	"time"

	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/subscription"
)

type SubscriptionGateway struct {
	secretKey string
}

func NewSubscriptionGateway(cfg *config.Config) *SubscriptionGateway {
	return &SubscriptionGateway{
		secretKey: cfg.StripeSecretKey,
	}
}

func (g *SubscriptionGateway) Retrieve(ctx context.Context, subscriptionID string) (*port.SubscriptionData, error) {
	if g.secretKey == "" {
		return nil, fmt.Errorf("stripe secret key is not configured")
	}

	stripe.Key = g.secretKey

	params := &stripe.SubscriptionParams{}
	params.Context = ctx

	sub, err := subscription.Get(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve subscription: %w", err)
	}

	return ExtractSubscriptionData(sub), nil
}

func (g *SubscriptionGateway) UpdatePrice(
	ctx context.Context,
	subscriptionID string,
	itemID string,
	priceID string,
	prorationDate *int64,
) (*port.SubscriptionData, error) {
	if g.secretKey == "" {
		return nil, fmt.Errorf("stripe secret key is not configured")
	}
	if subscriptionID == "" || itemID == "" || priceID == "" {
		return nil, fmt.Errorf("subscription id, item id and price id are required")
	}

	stripe.Key = g.secretKey

	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(priceID),
			},
		},
		ProrationBehavior: stripe.String("always_invoice"),
	}
	if prorationDate != nil && *prorationDate > 0 {
		params.ProrationDate = prorationDate
	}
	params.Context = ctx

	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription price: %w", err)
	}

	return ExtractSubscriptionData(sub), nil
}

func (g *SubscriptionGateway) PreviewPriceChange(
	ctx context.Context,
	subscriptionID string,
	itemID string,
	priceID string,
) (*port.ProrationPreview, error) {
	if g.secretKey == "" {
		return nil, fmt.Errorf("stripe secret key is not configured")
	}
	if subscriptionID == "" || itemID == "" || priceID == "" {
		return nil, fmt.Errorf("subscription id, item id and price id are required")
	}

	stripe.Key = g.secretKey

	prorationDate := time.Now().UTC().Unix()
	params := &stripe.InvoiceCreatePreviewParams{
		Subscription: stripe.String(subscriptionID),
		SubscriptionDetails: &stripe.InvoiceCreatePreviewSubscriptionDetailsParams{
			Items: []*stripe.InvoiceCreatePreviewSubscriptionDetailsItemParams{
				{
					ID:    stripe.String(itemID),
					Price: stripe.String(priceID),
				},
			},
			ProrationBehavior: stripe.String("always_invoice"),
			ProrationDate:     stripe.Int64(prorationDate),
		},
	}
	params.Context = ctx

	inv, err := invoice.CreatePreview(params)
	if err != nil {
		return nil, fmt.Errorf("failed to preview subscription price change: %w", err)
	}

	preview := &port.ProrationPreview{
		Currency:      string(inv.Currency),
		AmountDue:     inv.AmountDue,
		Subtotal:      inv.Subtotal,
		Total:         inv.Total,
		ProrationDate: prorationDate,
		PeriodStart:   unixToTime(inv.PeriodStart),
		PeriodEnd:     unixToTime(inv.PeriodEnd),
		Lines:         make([]port.ProrationPreviewLine, 0),
	}

	if inv.Lines != nil {
		for _, line := range inv.Lines.Data {
			if line == nil {
				continue
			}
			preview.Lines = append(preview.Lines, port.ProrationPreviewLine{
				Description: line.Description,
				Amount:      line.Amount,
				Proration:   lineIsProration(line),
			})
		}
	}

	return preview, nil
}

func lineIsProration(line *stripe.InvoiceLineItem) bool {
	if line == nil || line.Parent == nil || line.Parent.SubscriptionItemDetails == nil {
		return false
	}
	return line.Parent.SubscriptionItemDetails.Proration
}

func unixToTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}

func ExtractSubscriptionData(sub *stripe.Subscription) *port.SubscriptionData {
	data := &port.SubscriptionData{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}

	if sub.Customer != nil {
		data.CustomerID = sub.Customer.ID
	}

	if sub.Items != nil && len(sub.Items.Data) > 0 {
		item := sub.Items.Data[0]
		data.ItemID = item.ID
		if item.Price != nil {
			data.PriceID = item.Price.ID
		}
		if item.CurrentPeriodStart > 0 {
			data.CurrentPeriodStart = time.Unix(item.CurrentPeriodStart, 0).UTC()
		}
		if item.CurrentPeriodEnd > 0 {
			data.CurrentPeriodEnd = time.Unix(item.CurrentPeriodEnd, 0).UTC()
		}
	}

	return data
}
