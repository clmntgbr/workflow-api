package handler

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	endpointcmd "go-api/internal/application/command/endpoint"
	queryendpoint "go-api/internal/application/query/endpoint"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type EndpointHandler struct {
	createHandler    endpointCreateHandler
	importHandler    endpointImportHandler
	updateHandler    endpointUpdateHandler
	deleteHandler    endpointDeleteHandler
	getByIDHandler   endpointGetByIDHandler
	listByOrgHandler endpointListByProjectHandler
}

func NewEndpointHandler(
	createHandler endpointCreateHandler,
	importHandler endpointImportHandler,
	updateHandler endpointUpdateHandler,
	deleteHandler endpointDeleteHandler,
	getByIDHandler endpointGetByIDHandler,
	listByOrgHandler endpointListByProjectHandler,
) *EndpointHandler {
	return &EndpointHandler{
		createHandler:    createHandler,
		importHandler:    importHandler,
		updateHandler:    updateHandler,
		deleteHandler:    deleteHandler,
		getByIDHandler:   getByIDHandler,
		listByOrgHandler: listByOrgHandler,
	}
}

func (h *EndpointHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
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
	urlWithoutQuery, queryParams, err := httpquery.ResolveURLAndQuery(req.URL, req.Query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid URL"})
	}
	req.URL = urlWithoutQuery
	body := req.Body
	if body == nil {
		body = map[string]any{}
	}

	e, err := h.createHandler.Handle(c.Context(), endpointcmd.CreateEndpointCommand{
		UserID:         user.ID,
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Method:         method,
		Headers:        headers,
		Query:          queryParams,
		Body:           body,
		Timeout:        *req.Timeout,
		RetryOnFailure: *req.RetryOnFailure,
		RetryCount:     *req.RetryCount,
		RetryDelay:     *req.RetryDelay,
		ProjectID: orgID,
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create endpoint"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewEndpointDetailResponseFromEntity(*e))
}

const maxOpenAPIImportFileSize = 8 << 20

func (h *EndpointHandler) ImportFromOpenAPI(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil || fileHeader == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "OpenAPI file is required"})
	}
	if fileHeader.Size > maxOpenAPIImportFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "OpenAPI file is too large"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Failed to read OpenAPI file"})
	}
	defer file.Close()

	spec, err := io.ReadAll(io.LimitReader(file, maxOpenAPIImportFileSize+1))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Failed to read OpenAPI file"})
	}
	if int64(len(spec)) > maxOpenAPIImportFileSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "OpenAPI file is too large"})
	}

	payload := strings.TrimSpace(c.FormValue("payload"))
	if payload == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "payload is required"})
	}

	var req dto.ImportEndpointsFromOpenAPIRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid payload"})
	}
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.Status = strings.TrimSpace(req.Status)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	status, err := domainendpoint.ParseStatus(req.Status)
	if err != nil || status == domainendpoint.StatusDeleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid status"})
	}

	headers := req.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	queryParams := req.Query
	if queryParams == nil {
		queryParams = httpquery.Empty()
	}
	body := req.Body
	if body == nil {
		body = map[string]any{}
	}

	created, err := h.importHandler.Handle(c.Context(), endpointcmd.ImportEndpointsFromOpenAPICommand{
		UserID:         user.ID,
		Spec:           spec,
		BaseURL:        req.BaseURL,
		Status:         status,
		Headers:        headers,
		Query:          queryParams,
		Body:           body,
		Timeout:        *req.Timeout,
		RetryOnFailure: *req.RetryOnFailure,
		RetryCount:     *req.RetryCount,
		RetryDelay:     *req.RetryDelay,
		ProjectID: orgID,
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		if errors.Is(err, domainendpoint.ErrInvalidOpenAPI) ||
			errors.Is(err, domainendpoint.ErrNoOperations) ||
			errors.Is(err, domainendpoint.ErrTooManyOperations) ||
			errors.Is(err, domainendpoint.ErrInvalidEndpointURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to import endpoints"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewEndpointListResponseFromEntities(created))
}

func (h *EndpointHandler) Update(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid endpoint id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), queryendpoint.GetEndpointByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "endpoint not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get endpoint"})
	}
	if existing.ProjectID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
	}

	var req dto.UpdateEndpointRequest
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

	status, err := domainendpoint.ParseStatus(req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid status"})
	}

	headers := req.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	urlWithoutQuery, queryParams, err := httpquery.ResolveURLAndQuery(req.URL, req.Query)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid URL"})
	}
	req.URL = urlWithoutQuery
	body := req.Body
	if body == nil {
		body = map[string]any{}
	}

	err = h.updateHandler.Handle(c.Context(), endpointcmd.UpdateEndpointCommand{
		ID:             id,
		UserID:         user.ID,
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Method:         method,
		Headers:        headers,
		Query:          queryParams,
		Body:           body,
		Timeout:        *req.Timeout,
		RetryOnFailure: *req.RetryOnFailure,
		RetryCount:     *req.RetryCount,
		RetryDelay:     *req.RetryDelay,
		Status:         status,
	})
	if err != nil {
		if err.Error() == "endpoint not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
		}
		if err.Error() == "invalid status" || err.Error() == "use delete to mark an endpoint as deleted" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update endpoint"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryendpoint.GetEndpointByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get endpoint"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewEndpointDetailResponseFromView(*view))
}

func (h *EndpointHandler) GetByID(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
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
	if view.ProjectID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewEndpointDetailResponseFromView(*view))
}

func (h *EndpointHandler) ListByProject(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	var listQuery struct {
		paginate.PaginateQuery
		Method []string `query:"method"`
	}
	if err := c.Bind().Query(&listQuery); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}

	orderBy := listQuery.OrderBy
	sortBy := listQuery.SortBy
	listQuery.Normalize()
	if sortBy == "" {
		listQuery.SortBy = "created_at"
	}
	if orderBy == "" {
		listQuery.OrderBy = paginate.OrderByDesc
	}

	views, total, err := h.listByOrgHandler.Handle(c.Context(), queryendpoint.ListEndpointsByProjectQuery{
		ProjectID: orgID,
		Query:          listQuery.PaginateQuery,
		Methods:        listQuery.Method,
	})
	if err != nil {
		if strings.Contains(err.Error(), "invalid endpoint method") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list endpoints"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewEndpointListResponseFromViews(views),
		int(total),
		listQuery.PaginateQuery,
	))
}

func (h *EndpointHandler) Delete(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid endpoint id"})
	}

	if err := h.deleteHandler.Handle(c.Context(), endpointcmd.DeleteEndpointCommand{
		ID:        id,
		UserID:    user.ID,
		ProjectID: orgID,
	}); err != nil {
		if err.Error() == "endpoint not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Endpoint not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete endpoint"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
