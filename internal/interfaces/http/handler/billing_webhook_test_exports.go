package handler

import (
	"time"

	cmdsubscription "go-api/internal/application/command/subscription"

	"github.com/stripe/stripe-go/v82"
)

// UpsertInvoiceCommandFromStripeForTest exposes upsertInvoiceCommandFromStripe.
func UpsertInvoiceCommandFromStripeForTest(invoice *stripe.Invoice) cmdsubscription.UpsertInvoiceCommand {
	return upsertInvoiceCommandFromStripe(invoice)
}

// InvoiceDescriptionForTest exposes invoiceDescription.
func InvoiceDescriptionForTest(invoice *stripe.Invoice) string {
	return invoiceDescription(invoice)
}

// SubscriptionIDFromInvoiceForTest exposes subscriptionIDFromInvoice.
func SubscriptionIDFromInvoiceForTest(invoice *stripe.Invoice) string {
	return subscriptionIDFromInvoice(invoice)
}

// CustomerIDFromInvoiceForTest exposes customerIDFromInvoice.
func CustomerIDFromInvoiceForTest(invoice *stripe.Invoice) string {
	return customerIDFromInvoice(invoice)
}

// UnixToTimeForTest exposes unixToTime.
func UnixToTimeForTest(ts int64) time.Time {
	return unixToTime(ts)
}
