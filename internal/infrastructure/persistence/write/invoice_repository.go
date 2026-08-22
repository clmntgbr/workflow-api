package write

import (
	"context"
	"errors"

	domaininvoice "go-api/internal/domain/invoice"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type invoiceWriteRepository struct {
	db *gorm.DB
}

func NewInvoiceWriteRepository(db *gorm.DB) domaininvoice.InvoiceWriteRepository {
	return &invoiceWriteRepository{db: db}
}

func (r *invoiceWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *invoiceWriteRepository) GetByStripeInvoiceID(
	ctx context.Context,
	stripeInvoiceID string,
) (*domaininvoice.Invoice, error) {
	var model InvoiceModel
	err := DBWithContext(ctx, r.db).
		Where("stripe_invoice_id = ?", stripeInvoiceID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return invoiceDomainFromModel(&model), nil
}

func (r *invoiceWriteRepository) UpsertByStripeInvoiceID(ctx context.Context, invoice *domaininvoice.Invoice) error {
	existing, err := r.GetByStripeInvoiceID(ctx, invoice.StripeInvoiceID)
	if err != nil {
		return err
	}

	if existing == nil {
		return DBWithContext(ctx, r.db).Omit(clause.Associations).Create(invoiceModelFromDomain(invoice)).Error
	}

	invoice.ID = existing.ID
	invoice.CreatedAt = existing.CreatedAt
	return DBWithContext(ctx, r.db).Omit(clause.Associations).Save(invoiceModelFromDomain(invoice)).Error
}

func invoiceModelFromDomain(i *domaininvoice.Invoice) *InvoiceModel {
	return &InvoiceModel{
		ID:                   i.ID,
		UserID:               i.UserID,
		SubscriptionID:       i.SubscriptionID,
		StripeInvoiceID:      i.StripeInvoiceID,
		StripeCustomerID:     i.StripeCustomerID,
		StripeSubscriptionID: i.StripeSubscriptionID,
		Number:               i.Number,
		Status:               i.Status,
		Currency:             i.Currency,
		AmountDue:            i.AmountDue,
		AmountPaid:           i.AmountPaid,
		Total:                i.Total,
		HostedInvoiceURL:     i.HostedInvoiceURL,
		InvoicePDF:           i.InvoicePDF,
		BillingReason:        i.BillingReason,
		Description:          i.Description,
		AttemptCount:         i.AttemptCount,
		PeriodStart:          i.PeriodStart,
		PeriodEnd:            i.PeriodEnd,
		PaidAt:               i.PaidAt,
		StripeCreatedAt:      i.StripeCreatedAt,
		CreatedAt:            i.CreatedAt,
		UpdatedAt:            i.UpdatedAt,
	}
}

func invoiceDomainFromModel(model *InvoiceModel) *domaininvoice.Invoice {
	return &domaininvoice.Invoice{
		ID:                   model.ID,
		UserID:               model.UserID,
		SubscriptionID:       model.SubscriptionID,
		StripeInvoiceID:      model.StripeInvoiceID,
		StripeCustomerID:     model.StripeCustomerID,
		StripeSubscriptionID: model.StripeSubscriptionID,
		Number:               model.Number,
		Status:               model.Status,
		Currency:             model.Currency,
		AmountDue:            model.AmountDue,
		AmountPaid:           model.AmountPaid,
		Total:                model.Total,
		HostedInvoiceURL:     model.HostedInvoiceURL,
		InvoicePDF:           model.InvoicePDF,
		BillingReason:        model.BillingReason,
		Description:          model.Description,
		AttemptCount:         model.AttemptCount,
		PeriodStart:          model.PeriodStart,
		PeriodEnd:            model.PeriodEnd,
		PaidAt:               model.PaidAt,
		StripeCreatedAt:      model.StripeCreatedAt,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
}
