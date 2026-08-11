package handler

import (
	"strings"

	workflowcmd "go-api/internal/application/command/workflow"
	queryworkflow "go-api/internal/application/query/workflow"
	"go-api/internal/domain/paginate"
	domainworkflow "go-api/internal/domain/workflow"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type WorkflowHandler struct {
	createHandler    *workflowcmd.CreateWorkflowHandler
	updateHandler    *workflowcmd.UpdateWorkflowHandler
	deleteHandler    *workflowcmd.DeleteWorkflowHandler
	getByIDHandler   *queryworkflow.GetWorkflowByIDHandler
	listByOrgHandler *queryworkflow.ListWorkflowsByOrganizationHandler
}

func NewWorkflowHandler(
	createHandler *workflowcmd.CreateWorkflowHandler,
	updateHandler *workflowcmd.UpdateWorkflowHandler,
	deleteHandler *workflowcmd.DeleteWorkflowHandler,
	getByIDHandler *queryworkflow.GetWorkflowByIDHandler,
	listByOrgHandler *queryworkflow.ListWorkflowsByOrganizationHandler,
) *WorkflowHandler {
	return &WorkflowHandler{
		createHandler:    createHandler,
		updateHandler:    updateHandler,
		deleteHandler:    deleteHandler,
		getByIDHandler:   getByIDHandler,
		listByOrgHandler: listByOrgHandler,
	}
}

func (h *WorkflowHandler) Create(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	var req dto.CreateWorkflowRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	w, err := h.createHandler.Handle(c.Context(), workflowcmd.CreateWorkflowCommand{
		Name:                    req.Name,
		Description:             req.Description,
		OrganizationID:          orgID,
		ScheduleIntervalMinutes: intOrDefault(req.ScheduleIntervalMinutes, 0),
		Concurrency:             intOrDefault(req.Concurrency, 1),
		NotificationsEnabled:    boolOrDefault(req.NotificationsEnabled, true),
		NotifyOnSuccess:         boolOrDefault(req.NotifyOnSuccess, true),
		NotifyOnFailure:         boolOrDefault(req.NotifyOnFailure, true),
		NotifyOnCancel:          boolOrDefault(req.NotifyOnCancel, true),
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create workflow"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewWorkflowDetailResponseFromEntity(*w))
}

func (h *WorkflowHandler) GetByID(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if view.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromView(*view))
}

func (h *WorkflowHandler) ListByOrganization(c fiber.Ctx) error {
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

	views, total, err := h.listByOrgHandler.Handle(c.Context(), queryworkflow.ListWorkflowsByOrganizationQuery{
		OrganizationID: orgID,
		Query:          query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list workflows"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewWorkflowListResponseFromViews(views),
		int(total),
		query,
	))
}

func (h *WorkflowHandler) Update(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if existing.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	var req dto.UpdateWorkflowRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Description = strings.TrimSpace(req.Description)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	status, err := domainworkflow.ParseStatus(req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid status"})
	}

	err = h.updateHandler.Handle(c.Context(), workflowcmd.UpdateWorkflowCommand{
		ID:                      id,
		Name:                    req.Name,
		Description:             req.Description,
		Status:                  status,
		ScheduleIntervalMinutes: *req.ScheduleIntervalMinutes,
		Concurrency:             *req.Concurrency,
		NotificationsEnabled:    *req.NotificationsEnabled,
		NotifyOnSuccess:         *req.NotifyOnSuccess,
		NotifyOnFailure:         *req.NotifyOnFailure,
		NotifyOnCancel:          *req.NotifyOnCancel,
	})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if err.Error() == "invalid status" || err.Error() == "use delete to mark a workflow as deleted" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update workflow"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromView(*view))
}

func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if existing.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	if err := h.deleteHandler.Handle(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete workflow"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func boolOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func intOrDefault(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}
