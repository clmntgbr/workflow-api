package handler

import (
	"context"

	cmdsubscription "go-api/internal/application/command/subscription"
)

type billingCheckoutCompletedHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.CheckoutCompletedCommand) error
}

type billingSubscriptionUpdatedHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.SubscriptionUpdatedCommand) error
}

type billingSubscriptionDeletedHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.SubscriptionDeletedCommand) error
}

type billingInvoicePaymentSucceededHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.InvoicePaymentSucceededCommand) error
}

type billingInvoicePaymentFailedHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.InvoicePaymentFailedCommand) error
}

type billingUpsertInvoiceHandler interface {
	Handle(ctx context.Context, cmd cmdsubscription.UpsertInvoiceCommand) error
}
