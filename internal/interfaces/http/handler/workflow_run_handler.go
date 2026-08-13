package handler

import (
	"errors"

	workflowruncmd "go-api/internal/application/command/workflowrun"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type WorkflowRunHandler struct {
	startHandler       *workflowruncmd.StartWorkflowRunHandler
	getByIDHandler     *queryworkflowrun.GetWorkflowRunByIDHandler
	listByWorkflow     *queryworkflowrun.ListWorkflowRunsByWorkflowHandler
	listStepRuns       *querysteprun.ListStepRunsByWorkflowRunHandler
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler
}

func NewWorkflowRunHandler(
	startHandler *workflowruncmd.StartWorkflowRunHandler,
	getByIDHandler *queryworkflowrun.GetWorkflowRunByIDHandler,
	listByWorkflow *queryworkflowrun.ListWorkflowRunsByWorkflowHandler,
	listStepRuns *querysteprun.ListStepRunsByWorkflowRunHandler,
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler,
) *WorkflowRunHandler {
	return &WorkflowRunHandler{
		startHandler:       startHandler,
		getByIDHandler:     getByIDHandler,
		listByWorkflow:     listByWorkflow,
		listStepRuns:       listStepRuns,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *WorkflowRunHandler) Start(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
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
	if workflow.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	var req dto.StartWorkflowRunRequest
	if len(c.Body()) > 0 {
		if err := validation.BindBody(c, &req); err != nil {
			return err
		}
	}

	userID := user.ID
	run, err := h.startHandler.Handle(c.Context(), workflowruncmd.StartWorkflowRunCommand{
		WorkflowID:        workflowID,
		TriggeredBy:       domainworkflowrun.TriggeredByUser,
		TriggeredByUserID: &userID,
		Context:           req.Context,
	})
	if err != nil {
		if errors.Is(err, domainworkflowrun.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if errors.Is(err, domainworkflowrun.ErrAlreadyInProgress) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "RUN_IN_PROGRESS",
				"message": "A workflow run is already in progress",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to start workflow run"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewWorkflowRunDetailResponseFromEntity(*run, orgID))
}

func (h *WorkflowRunHandler) ListByWorkflow(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
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
	if workflow.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
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

	views, total, err := h.listByWorkflow.Handle(c.Context(), queryworkflowrun.ListWorkflowRunsByWorkflowQuery{
		WorkflowID: workflowID,
		Query:      query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list workflow runs"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewWorkflowRunListResponseFromViews(views),
		int(total),
		query,
	))
}

func (h *WorkflowRunHandler) GetByID(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow run id"})
	}

	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if workflow.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflowrun.GetWorkflowRunByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow run not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow run not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow run"})
	}
	if view.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow run not found"})
	}

	stepRuns, err := h.listStepRuns.Handle(c.Context(), querysteprun.ListStepRunsByWorkflowRunQuery{
		WorkflowRunID: view.ID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list step runs"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowRunDetailResponseFromView(*view, stepRuns))
}
