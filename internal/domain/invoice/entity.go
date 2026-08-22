package invoice

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Invoice struct {
	ID uuid.UUID

	UserID         uuid.UUID
	SubscriptionID *uuid.UUID

	StripeInvoiceID      string
	StripeCustomerID     string
	StripeSubscriptionID string

	Number   string
	Status   string
	Currency string

	AmountDue  int64
	AmountPaid int64
	Total      int64

	HostedInvoiceURL string
	InvoicePDF       string

	BillingReason string
	Description   string
	AttemptCount  int64

	PeriodStart     time.Time
	PeriodEnd       time.Time
	PaidAt          *time.Time
	StripeCreatedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewInvoice(userID uuid.UUID, subscriptionID *uuid.UUID, stripeInvoiceID string) *Invoice {
	now := time.Now().UTC()
	return &Invoice{
		ID:              uuid.New(),
		UserID:          userID,
		SubscriptionID:  subscriptionID,
		StripeInvoiceID: stripeInvoiceID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (i *Invoice) PullEvents() []event.DomainEvent {
	events := i.events
	i.events = nil
	return events
}

func (i *Invoice) recordEvent(e event.DomainEvent) {
	i.events = append(i.events, e)
}

func (i *Invoice) ApplyStripeSnapshot(
	stripeCustomerID string,
	stripeSubscriptionID string,
	number string,
	status string,
	currency string,
	amountDue int64,
	amountPaid int64,
	total int64,
	hostedInvoiceURL string,
	invoicePDF string,
	billingReason string,
	description string,
	attemptCount int64,
	periodStart time.Time,
	periodEnd time.Time,
	paidAt *time.Time,
	stripeCreatedAt time.Time,
) {
	i.StripeCustomerID = stripeCustomerID
	i.StripeSubscriptionID = stripeSubscriptionID
	i.Number = number
	i.Status = status
	i.Currency = currency
	i.AmountDue = amountDue
	i.AmountPaid = amountPaid
	i.Total = total
	i.HostedInvoiceURL = hostedInvoiceURL
	i.InvoicePDF = invoicePDF
	i.BillingReason = billingReason
	i.Description = description
	i.AttemptCount = attemptCount
	i.PeriodStart = periodStart
	i.PeriodEnd = periodEnd
	i.PaidAt = paidAt
	i.StripeCreatedAt = stripeCreatedAt
	i.UpdatedAt = time.Now().UTC()
}

func (i *Invoice) RaiseCreated() {
	i.recordEvent(InvoiceCreated{
		ID:                   uuid.New().String(),
		InvoiceID:            i.ID.String(),
		UserID:               i.UserID.String(),
		SubscriptionID:       subscriptionIDString(i.SubscriptionID),
		StripeInvoiceID:      i.StripeInvoiceID,
		StripeCustomerID:     i.StripeCustomerID,
		StripeSubscriptionID: i.StripeSubscriptionID,
		Number:               i.Number,
		Status:               i.Status,
		Currency:             i.Currency,
		AmountDue:            i.AmountDue,
		AmountPaid:           i.AmountPaid,
		Total:                i.Total,
		BillingReason:        i.BillingReason,
		PaidAt:               i.PaidAt,
		StripeCreatedAt:      i.StripeCreatedAt,
		Timestamp:            i.UpdatedAt,
	})
}

func (i *Invoice) RaiseUpdated() {
	i.recordEvent(InvoiceUpdated{
		ID:                   uuid.New().String(),
		InvoiceID:            i.ID.String(),
		UserID:               i.UserID.String(),
		SubscriptionID:       subscriptionIDString(i.SubscriptionID),
		StripeInvoiceID:      i.StripeInvoiceID,
		StripeCustomerID:     i.StripeCustomerID,
		StripeSubscriptionID: i.StripeSubscriptionID,
		Number:               i.Number,
		Status:               i.Status,
		Currency:             i.Currency,
		AmountDue:            i.AmountDue,
		AmountPaid:           i.AmountPaid,
		Total:                i.Total,
		BillingReason:        i.BillingReason,
		PaidAt:               i.PaidAt,
		StripeCreatedAt:      i.StripeCreatedAt,
		Timestamp:            i.UpdatedAt,
	})
}

func subscriptionIDString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}
