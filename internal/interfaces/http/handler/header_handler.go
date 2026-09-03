package handler

import (
	queryheader "go-api/internal/application/query/header"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
)

type HeaderHandler struct {
	suggestHandler      headerSuggestHandler
	suggestValueHandler headerValueSuggestHandler
}

func NewHeaderHandler(
	suggestHandler headerSuggestHandler,
	suggestValueHandler headerValueSuggestHandler,
) *HeaderHandler {
	return &HeaderHandler{
		suggestHandler:      suggestHandler,
		suggestValueHandler: suggestValueHandler,
	}
}

func (h *HeaderHandler) Suggest(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
		})
	}
	query.Normalize()

	suggestions, total, err := h.suggestHandler.Handle(c.Context(), queryheader.SuggestHeadersQuery{
		ProjectID: projectID,
		Search:    query.Search,
		Paginate:  query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch header suggestions"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewHeaderSuggestionsResponse(suggestions),
		int(total),
		query,
	))
}

func (h *HeaderHandler) SuggestValues(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
		})
	}
	query.Normalize()

	suggestions, total, err := h.suggestValueHandler.Handle(c.Context(), queryheader.SuggestHeaderValuesQuery{
		ProjectID: projectID,
		Key:       c.Query("key", ""),
		Search:    query.Search,
		Paginate:  query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch header value suggestions"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewHeaderValueSuggestionsResponse(suggestions),
		int(total),
		query,
	))
}
