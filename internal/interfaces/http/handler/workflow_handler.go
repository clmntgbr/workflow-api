package handler

import (
	"strings"

	workflowcmd "go-api/internal/application/command/workflow"
	queryworkflow "go-api/internal/application/query/workflow"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"

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
	var req dto.CreateWorkflowRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "name is required"})
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid organization id"})
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

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromView(*view))
}

func (h *WorkflowHandler) ListByOrganization(c fiber.Ctx) error {
	orgID, err := uuid.Parse(c.Query("organizationId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "organizationId query param is required"})
	}

	views, err := h.listByOrgHandler.Handle(c.Context(), queryworkflow.ListWorkflowsByOrganizationQuery{
		OrganizationID: orgID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list workflows"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowListResponseFromViews(views))
}

func (h *WorkflowHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	var req dto.UpdateWorkflowRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "name is required"})
	}

	status, err := domainworkflow.ParseStatus(req.Status)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid status"})
	}
	if status == domainworkflow.StatusDeleted {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Use DELETE to mark a workflow as deleted"})
	}

	if req.NotificationsEnabled == nil || req.NotifyOnSuccess == nil ||
		req.NotifyOnFailure == nil || req.NotifyOnCancel == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "notification flags are required"})
	}
	if req.ScheduleIntervalMinutes == nil || req.Concurrency == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "scheduleIntervalMinutes and concurrency are required"})
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
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
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
