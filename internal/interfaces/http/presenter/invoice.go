package presenter

import (
	"time"

	domaininvoice "go-api/internal/domain/invoice"
)

type InvoiceResponse struct {
	ID               string     `json:"id"`
	SubscriptionID   *string    `json:"subscriptionId"`
	StripeInvoiceID  *string    `json:"stripeInvoiceId"`
	Number           *string    `json:"number"`
	Status           *string    `json:"status"`
	Currency         *string    `json:"currency"`
	AmountDue        int64      `json:"amountDue"`
	AmountPaid       int64      `json:"amountPaid"`
	Total            int64      `json:"total"`
	HostedInvoiceURL *string    `json:"hostedInvoiceUrl"`
	InvoicePDF       *string    `json:"invoicePdf"`
	BillingReason    *string    `json:"billingReason"`
	Description      *string    `json:"description"`
	AttemptCount     int64      `json:"attemptCount"`
	PeriodStart      time.Time  `json:"periodStart"`
	PeriodEnd        time.Time  `json:"periodEnd"`
	PaidAt           *time.Time `json:"paidAt"`
	StripeCreatedAt  time.Time  `json:"stripeCreatedAt"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

func NewInvoiceResponse(invoice *domaininvoice.InvoiceView) InvoiceResponse {
	var subscriptionID *string
	if invoice.SubscriptionID != nil {
		value := invoice.SubscriptionID.String()
		subscriptionID = &value
	}

	return InvoiceResponse{
		ID:               invoice.ID.String(),
		SubscriptionID:   subscriptionID,
		StripeInvoiceID:  optionalNonEmptyString(invoice.StripeInvoiceID),
		Number:           optionalNonEmptyString(invoice.Number),
		Status:           optionalNonEmptyString(invoice.Status),
		Currency:         optionalNonEmptyString(invoice.Currency),
		AmountDue:        invoice.AmountDue,
		AmountPaid:       invoice.AmountPaid,
		Total:            invoice.Total,
		HostedInvoiceURL: optionalNonEmptyString(invoice.HostedInvoiceURL),
		InvoicePDF:       optionalNonEmptyString(invoice.InvoicePDF),
		BillingReason:    optionalNonEmptyString(invoice.BillingReason),
		Description:      optionalNonEmptyString(invoice.Description),
		AttemptCount:     invoice.AttemptCount,
		PeriodStart:      invoice.PeriodStart,
		PeriodEnd:        invoice.PeriodEnd,
		PaidAt:           invoice.PaidAt,
		StripeCreatedAt:  invoice.StripeCreatedAt,
		CreatedAt:        invoice.CreatedAt,
		UpdatedAt:        invoice.UpdatedAt,
	}
}

func NewInvoiceResponses(invoices []*domaininvoice.InvoiceView) []InvoiceResponse {
	responses := make([]InvoiceResponse, 0, len(invoices))
	for _, invoice := range invoices {
		if invoice == nil {
			continue
		}
		responses = append(responses, NewInvoiceResponse(invoice))
	}
	return responses
}
