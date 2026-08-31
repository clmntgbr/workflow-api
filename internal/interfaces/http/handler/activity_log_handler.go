package handler

import (
	queryactivitylog "go-api/internal/application/query/activitylog"
	queryworkflow "go-api/internal/application/query/workflow"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/presenter"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ActivityLogHandler struct {
	listByWorkflow     activityLogListByWorkflowHandler
	getWorkflowHandler activityLogGetWorkflowHandler
}

func NewActivityLogHandler(
	listByWorkflow activityLogListByWorkflowHandler,
	getWorkflowHandler activityLogGetWorkflowHandler,
) *ActivityLogHandler {
	return &ActivityLogHandler{
		listByWorkflow:     listByWorkflow,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *ActivityLogHandler) ListByWorkflow(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if workflow.ProjectID != projectID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}

	views, total, err := h.listByWorkflow.Handle(c.Context(), queryactivitylog.ListByWorkflowQuery{
		UserID:     user.ID,
		ProjectID:  projectID,
		WorkflowID: workflowID,
		Query:      query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list activity logs"})
	}

	query.Normalize()
	return c.JSON(presenter.NewActivityLogPaginateResponse(views, total, query))
}
