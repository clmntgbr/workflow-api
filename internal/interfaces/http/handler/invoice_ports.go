package handler

import (
	"context"

	queryinvoice "go-api/internal/application/query/invoice"
	domaininvoice "go-api/internal/domain/invoice"
)

type invoiceListHandler interface {
	Handle(ctx context.Context, q queryinvoice.ListInvoicesQuery) ([]*domaininvoice.InvoiceView, int64, error)
}
