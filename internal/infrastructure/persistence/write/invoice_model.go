package write

import (
	"time"

	"github.com/google/uuid"
)

type InvoiceModel struct {
	ID                   uuid.UUID  `gorm:"column:id;primaryKey"`
	UserID               uuid.UUID  `gorm:"column:user_id"`
	SubscriptionID       *uuid.UUID `gorm:"column:subscription_id"`
	StripeInvoiceID      string     `gorm:"column:stripe_invoice_id"`
	StripeCustomerID     string     `gorm:"column:stripe_customer_id"`
	StripeSubscriptionID string     `gorm:"column:stripe_subscription_id"`
	Number               string     `gorm:"column:number"`
	Status               string     `gorm:"column:status"`
	Currency             string     `gorm:"column:currency"`
	AmountDue            int64      `gorm:"column:amount_due"`
	AmountPaid           int64      `gorm:"column:amount_paid"`
	Total                int64      `gorm:"column:total"`
	HostedInvoiceURL     string     `gorm:"column:hosted_invoice_url"`
	InvoicePDF           string     `gorm:"column:invoice_pdf"`
	BillingReason        string     `gorm:"column:billing_reason"`
	Description          string     `gorm:"column:description"`
	AttemptCount         int64      `gorm:"column:attempt_count"`
	PeriodStart          time.Time  `gorm:"column:period_start"`
	PeriodEnd            time.Time  `gorm:"column:period_end"`
	PaidAt               *time.Time `gorm:"column:paid_at"`
	StripeCreatedAt      time.Time  `gorm:"column:stripe_created_at"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
}

func (InvoiceModel) TableName() string {
	return "invoices"
}
