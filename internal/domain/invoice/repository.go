package invoice

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type InvoiceWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	GetByStripeInvoiceID(ctx context.Context, stripeInvoiceID string) (*Invoice, error)
	UpsertByStripeInvoiceID(ctx context.Context, invoice *Invoice) error
}

type InvoiceReadRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID, query paginate.PaginateQuery) ([]*InvoiceView, int64, error)
}

type InvoiceView struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	SubscriptionID       *uuid.UUID
	StripeInvoiceID      string
	StripeCustomerID     string
	StripeSubscriptionID string
	Number               string
	Status               string
	Currency             string
	AmountDue            int64
	AmountPaid           int64
	Total                int64
	HostedInvoiceURL     string
	InvoicePDF           string
	BillingReason        string
	Description          string
	AttemptCount         int64
	PeriodStart          time.Time
	PeriodEnd            time.Time
	PaidAt               *time.Time
	StripeCreatedAt      time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
