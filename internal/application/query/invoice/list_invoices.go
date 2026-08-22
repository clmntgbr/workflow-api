package invoice

import (
	"context"
	"errors"

	domaininvoice "go-api/internal/domain/invoice"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ListInvoicesQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
}

type ListInvoicesHandler struct {
	invoiceRepo domaininvoice.InvoiceReadRepository
}

func NewListInvoicesHandler(invoiceRepo domaininvoice.InvoiceReadRepository) *ListInvoicesHandler {
	return &ListInvoicesHandler{invoiceRepo: invoiceRepo}
}

func (h *ListInvoicesHandler) Handle(
	ctx context.Context,
	q ListInvoicesQuery,
) ([]*domaininvoice.InvoiceView, int64, error) {
	invoices, total, err := h.invoiceRepo.FindByUserID(ctx, q.UserID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list invoices")
	}
	return invoices, total, nil
}
