package handler

import (
	queryinvoice "go-api/internal/application/query/invoice"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type InvoiceHandler struct {
	listInvoicesHandler *queryinvoice.ListInvoicesHandler
}

func NewInvoiceHandler(listInvoicesHandler *queryinvoice.ListInvoicesHandler) *InvoiceHandler {
	return &InvoiceHandler{listInvoicesHandler: listInvoicesHandler}
}

func (h *InvoiceHandler) GetInvoices(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}

	orderBy := query.OrderBy
	sortBy := query.SortBy
	query.Normalize()
	if sortBy == "" {
		query.SortBy = "stripe_created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	invoices, total, err := h.listInvoicesHandler.Handle(c.Context(), queryinvoice.ListInvoicesQuery{
		UserID: user.ID,
		Query:  query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to list invoices",
		})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewInvoiceResponses(invoices),
		int(total),
		query,
	))
}
