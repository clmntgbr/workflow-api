package handler

import (
	"strings"

	endpointcmd "go-api/internal/application/command/endpoint"
	queryendpoint "go-api/internal/application/query/endpoint"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type EndpointHandler struct {
	createHandler    *endpointcmd.CreateEndpointHandler
	getByIDHandler   *queryendpoint.GetEndpointByIDHandler
	listByOrgHandler *queryendpoint.ListEndpointsByOrganizationHandler
}

func NewEndpointHandler(
	createHandler *endpointcmd.CreateEndpointHandler,
	getByIDHandler *queryendpoint.GetEndpointByIDHandler,
	listByOrgHandler *queryendpoint.ListEndpointsByOrganizationHandler,
) *EndpointHandler {
	return &EndpointHandler{
		createHandler:    createHandler,
		getByIDHandler:   getByIDHandler,
		listByOrgHandler: listByOrgHandler,
	}
}

func (h *EndpointHandler) Create(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	var req dto.CreateEndpointRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	req.URL = strings.TrimSpace(req.URL)
	req.Method = strings.TrimSpace(req.Method)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	method, err := domainendpoint.ParseMethod(req.Method)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid method"})
	}

	headers := req.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	queryParams := req.Query
	if queryParams == nil {
		queryParams = map[string]string{}
	}

	e, err := h.createHandler.Handle(c.Context(), endpointcmd.CreateEndpointCommand{
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Method:         method,
		Headers:        headers,
		Query:          queryParams,
		Timeout:        intOrDefault(req.Timeout, 30000),
		RetryOnFailure: boolOrDefault(req.RetryOnFailure, false),
		RetryCount:     intOrDefault(req.RetryCount, 0),
		RetryDelay:     intOrDefault(req.RetryDelay, 1000),
		OrganizationID: orgID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create endpoint"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewEndpointDetailResponseFromEntity(*e))
}

func (h *EndpointHandler) GetByID(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid endpoint id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryendpoint.GetEndpointByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "endpoint not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get endpoint"})
	}
	if view.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewEndpointDetailResponseFromView(*view))
}

func (h *EndpointHandler) ListByOrganization(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
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
		query.SortBy = "created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	views, total, err := h.listByOrgHandler.Handle(c.Context(), queryendpoint.ListEndpointsByOrganizationQuery{
		OrganizationID: orgID,
		Query:          query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list endpoints"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewEndpointListResponseFromViews(views),
		int(total),
		query,
	))
}
