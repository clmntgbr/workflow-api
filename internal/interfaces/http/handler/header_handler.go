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

	var queryParams struct {
		paginate.PaginateQuery
		Search string `query:"search"`
	}
	if err := c.Bind().Query(&queryParams); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	queryParams.Normalize()

	suggestions, total, err := h.suggestHandler.Handle(c.Context(), queryheader.SuggestHeadersQuery{
		ProjectID: projectID,
		Search:    queryParams.Search,
		Paginate:  queryParams.PaginateQuery,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch header suggestions"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"items": presenter.NewHeaderSuggestionsResponse(suggestions),
			"page":  queryParams.Page,
			"limit": queryParams.Limit,
			"total": total,
		},
	})
}

func (h *HeaderHandler) SuggestValues(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	key := c.Query("key", "") // Optional now

	var queryParams struct {
		paginate.PaginateQuery
		Search string `query:"search"`
	}
	if err := c.Bind().Query(&queryParams); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	queryParams.Normalize()

	suggestions, total, err := h.suggestValueHandler.Handle(c.Context(), queryheader.SuggestHeaderValuesQuery{
		ProjectID: projectID,
		Key:       key, // Can be empty to search all
		Search:    queryParams.Search,
		Paginate:  queryParams.PaginateQuery,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to fetch header value suggestions"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"items": presenter.NewHeaderValueSuggestionsResponse(suggestions),
			"page":  queryParams.Page,
			"limit": queryParams.Limit,
			"total": total,
		},
	})
}
