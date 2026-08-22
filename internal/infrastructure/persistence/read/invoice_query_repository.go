package read

import (
	"context"
	"time"

	domaininvoice "go-api/internal/domain/invoice"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type invoiceRow struct {
	ID                   uuid.UUID  `gorm:"column:id"`
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

func (invoiceRow) TableName() string { return "invoices" }

type invoiceReadRepository struct {
	db *gorm.DB
}

func NewInvoiceReadRepository(db *gorm.DB) domaininvoice.InvoiceReadRepository {
	return &invoiceReadRepository{db: db}
}

func (r *invoiceReadRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
	query paginate.PaginateQuery,
) ([]*domaininvoice.InvoiceView, int64, error) {
	if query.SortBy == "" {
		query.SortBy = "stripe_created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).Model(&invoiceRow{}).Where("user_id = ?", userID)
	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []invoiceRow
	if err := db.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	views := make([]*domaininvoice.InvoiceView, 0, len(rows))
	for i := range rows {
		views = append(views, invoiceViewFromRow(&rows[i]))
	}
	return views, total, nil
}

func invoiceViewFromRow(row *invoiceRow) *domaininvoice.InvoiceView {
	return &domaininvoice.InvoiceView{
		ID:                   row.ID,
		UserID:               row.UserID,
		SubscriptionID:       row.SubscriptionID,
		StripeInvoiceID:      row.StripeInvoiceID,
		StripeCustomerID:     row.StripeCustomerID,
		StripeSubscriptionID: row.StripeSubscriptionID,
		Number:               row.Number,
		Status:               row.Status,
		Currency:             row.Currency,
		AmountDue:            row.AmountDue,
		AmountPaid:           row.AmountPaid,
		Total:                row.Total,
		HostedInvoiceURL:     row.HostedInvoiceURL,
		InvoicePDF:           row.InvoicePDF,
		BillingReason:        row.BillingReason,
		Description:          row.Description,
		AttemptCount:         row.AttemptCount,
		PeriodStart:          row.PeriodStart,
		PeriodEnd:            row.PeriodEnd,
		PaidAt:               row.PaidAt,
		StripeCreatedAt:      row.StripeCreatedAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}
