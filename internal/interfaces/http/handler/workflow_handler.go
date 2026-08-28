package handler

import (
	"errors"
	"strings"

	workflowcmd "go-api/internal/application/command/workflow"
	queryproject "go-api/internal/application/query/project"
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
	createHandler     *workflowcmd.CreateWorkflowHandler
	updateHandler     *workflowcmd.UpdateWorkflowHandler
	activateHandler   *workflowcmd.ActivateWorkflowHandler
	deactivateHandler *workflowcmd.DeactivateWorkflowHandler
	deleteHandler     *workflowcmd.DeleteWorkflowHandler
	getByIDHandler    *queryworkflow.GetWorkflowByIDHandler
	listByOrgHandler  *queryworkflow.ListWorkflowsByProjectHandler
	getProjectByIDHandler *queryproject.GetProjectByIDHandler
}

func NewWorkflowHandler(
	createHandler *workflowcmd.CreateWorkflowHandler,
	updateHandler *workflowcmd.UpdateWorkflowHandler,
	activateHandler *workflowcmd.ActivateWorkflowHandler,
	deactivateHandler *workflowcmd.DeactivateWorkflowHandler,
	deleteHandler *workflowcmd.DeleteWorkflowHandler,
	getByIDHandler *queryworkflow.GetWorkflowByIDHandler,
	listByOrgHandler *queryworkflow.ListWorkflowsByProjectHandler,
	getProjectByIDHandler *queryproject.GetProjectByIDHandler,
) *WorkflowHandler {
	return &WorkflowHandler{
		createHandler:     createHandler,
		updateHandler:     updateHandler,
		activateHandler:   activateHandler,
		deactivateHandler: deactivateHandler,
		deleteHandler:     deleteHandler,
		getByIDHandler:    getByIDHandler,
		listByOrgHandler:  listByOrgHandler,
		getProjectByIDHandler: getProjectByIDHandler,
	}
}

func (h *WorkflowHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
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
		UserID:                user.ID,
		Name:                  req.Name,
		Description:           req.Description,
		ProjectID:        orgID,
		ScheduleType:          parseScheduleTypeOrNone(req.ScheduleType),
		ScheduleIntervalValue: intOrDefault(req.ScheduleIntervalValue, 0),
		ScheduleIntervalUnit:  domainworkflow.ScheduleUnit(req.ScheduleIntervalUnit),
		ScheduleAt:            req.ScheduleAt,
		ScheduleTimezone:      strings.TrimSpace(req.ScheduleTimezone),
		Concurrency:           intOrDefault(req.Concurrency, 1),
		NotificationsEnabled:  boolOrDefault(req.NotificationsEnabled, true),
		NotifyOnSuccess:       boolOrDefault(req.NotifyOnSuccess, true),
		NotifyOnFailure:       boolOrDefault(req.NotifyOnFailure, true),
		NotifyOnCancel:        boolOrDefault(req.NotifyOnCancel, true),
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		if status, message := scheduleError(err); status != 0 {
			return c.Status(status).JSON(fiber.Map{"message": message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create workflow"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewWorkflowDetailResponseFromEntity(*w))
}

func (h *WorkflowHandler) GetByID(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}

	if view.ProjectID == orgID {
		return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromView(*view))
	}

	org, err := h.getProjectByIDHandler.Handle(c.Context(), queryproject.GetProjectByIDQuery{ID: view.ProjectID})
	if err != nil || org == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	isMember := false
	for _, memberID := range org.MemberIDs {
		if memberID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	return c.Status(fiber.StatusConflict).JSON(fiber.Map{
		"code":             "WRONG_ORGANIZATION",
		"message":          "Workflow belongs to another project",
		"projectId":   org.ID.String(),
		"projectName": org.Name,
	})
}

func (h *WorkflowHandler) ListByProject(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
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

	views, total, err := h.listByOrgHandler.Handle(c.Context(), queryworkflow.ListWorkflowsByProjectQuery{
		ProjectID: orgID,
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if existing.ProjectID != orgID {
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
		ID:                    id,
		UserID:                user.ID,
		Name:                  req.Name,
		Description:           req.Description,
		Status:                status,
		ScheduleType:          parseScheduleTypeOrNone(req.ScheduleType),
		ScheduleIntervalValue: intOrDefault(req.ScheduleIntervalValue, 0),
		ScheduleIntervalUnit:  domainworkflow.ScheduleUnit(req.ScheduleIntervalUnit),
		ScheduleAt:            req.ScheduleAt,
		ScheduleTimezone:      strings.TrimSpace(req.ScheduleTimezone),
		Concurrency:           intOrDefault(req.Concurrency, existing.Concurrency),
		NotificationsEnabled:  boolOrDefault(req.NotificationsEnabled, existing.NotificationsEnabled),
		NotifyOnSuccess:       boolOrDefault(req.NotifyOnSuccess, existing.NotifyOnSuccess),
		NotifyOnFailure:       boolOrDefault(req.NotifyOnFailure, existing.NotifyOnFailure),
		NotifyOnCancel:        boolOrDefault(req.NotifyOnCancel, existing.NotifyOnCancel),
	})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if err.Error() == "invalid status" || err.Error() == "use delete to mark a workflow as deleted" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		if status, message := scheduleError(err); status != 0 {
			return c.Status(status).JSON(fiber.Map{"message": message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update workflow"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromView(*view))
}

func (h *WorkflowHandler) Activate(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	w, err := h.activateHandler.Handle(c.Context(), workflowcmd.ActivateWorkflowCommand{
		ID:        id,
		UserID:    user.ID,
		ProjectID: orgID,
	})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if errors.Is(err, domainworkflow.ErrInvalidStatusTransition) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to activate workflow"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromEntity(*w))
}

func (h *WorkflowHandler) Deactivate(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	w, err := h.deactivateHandler.Handle(c.Context(), workflowcmd.DeactivateWorkflowCommand{
		ID:        id,
		UserID:    user.ID,
		ProjectID: orgID,
	})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if errors.Is(err, domainworkflow.ErrInvalidStatusTransition) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to deactivate workflow"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowDetailResponseFromEntity(*w))
}

func (h *WorkflowHandler) Delete(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if existing.ProjectID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	if err := h.deleteHandler.Handle(c.Context(), workflowcmd.DeleteWorkflowCommand{
		ID:     id,
		UserID: user.ID,
	}); err != nil {
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

func parseScheduleTypeOrNone(value string) domainworkflow.ScheduleType {
	if value == "" {
		return domainworkflow.ScheduleTypeNone
	}
	return domainworkflow.ScheduleType(value)
}

func scheduleError(err error) (int, string) {
	switch {
	case errors.Is(err, domainworkflow.ErrScheduleIntervalTooShort):
		return fiber.StatusBadRequest, err.Error()
	case errors.Is(err, domainworkflow.ErrInvalidSchedule), errors.Is(err, domainworkflow.ErrInvalidScheduleTimezone):
		return fiber.StatusBadRequest, err.Error()
	default:
		return 0, ""
	}
}
