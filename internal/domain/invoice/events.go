package invoice

import (
	"time"

	"go-api/internal/domain/event"
)

const (
	EventTypeInvoiceCreated = "invoice.created.v1"
	EventTypeInvoiceUpdated = "invoice.updated.v1"
)

type InvoiceCreated struct {
	ID                   string     `json:"eventId"`
	InvoiceID            string     `json:"invoiceId"`
	UserID               string     `json:"userId"`
	SubscriptionID       *string    `json:"subscriptionId,omitempty"`
	StripeInvoiceID      string     `json:"stripeInvoiceId"`
	StripeCustomerID     string     `json:"stripeCustomerId"`
	StripeSubscriptionID string     `json:"stripeSubscriptionId"`
	Number               string     `json:"number"`
	Status               string     `json:"status"`
	Currency             string     `json:"currency"`
	AmountDue            int64      `json:"amountDue"`
	AmountPaid           int64      `json:"amountPaid"`
	Total                int64      `json:"total"`
	BillingReason        string     `json:"billingReason"`
	PaidAt               *time.Time `json:"paidAt,omitempty"`
	StripeCreatedAt      time.Time  `json:"stripeCreatedAt"`
	Timestamp            time.Time  `json:"timestamp"`
}

func (e InvoiceCreated) EventID() string       { return e.ID }
func (e InvoiceCreated) EventType() string     { return EventTypeInvoiceCreated }
func (e InvoiceCreated) AggregateID() string   { return e.InvoiceID }
func (e InvoiceCreated) OccurredAt() time.Time { return e.Timestamp }

type InvoiceUpdated struct {
	ID                   string     `json:"eventId"`
	InvoiceID            string     `json:"invoiceId"`
	UserID               string     `json:"userId"`
	SubscriptionID       *string    `json:"subscriptionId,omitempty"`
	StripeInvoiceID      string     `json:"stripeInvoiceId"`
	StripeCustomerID     string     `json:"stripeCustomerId"`
	StripeSubscriptionID string     `json:"stripeSubscriptionId"`
	Number               string     `json:"number"`
	Status               string     `json:"status"`
	Currency             string     `json:"currency"`
	AmountDue            int64      `json:"amountDue"`
	AmountPaid           int64      `json:"amountPaid"`
	Total                int64      `json:"total"`
	BillingReason        string     `json:"billingReason"`
	PaidAt               *time.Time `json:"paidAt,omitempty"`
	StripeCreatedAt      time.Time  `json:"stripeCreatedAt"`
	Timestamp            time.Time  `json:"timestamp"`
}

func (e InvoiceUpdated) EventID() string       { return e.ID }
func (e InvoiceUpdated) EventType() string     { return EventTypeInvoiceUpdated }
func (e InvoiceUpdated) AggregateID() string   { return e.InvoiceID }
func (e InvoiceUpdated) OccurredAt() time.Time { return e.Timestamp }

var (
	_ event.DomainEvent = InvoiceCreated{}
	_ event.DomainEvent = InvoiceUpdated{}
)
