package handler

import (
	"strings"

	stepcmd "go-api/internal/application/command/step"
	querystep "go-api/internal/application/query/step"
	queryworkflow "go-api/internal/application/query/workflow"
	domainstep "go-api/internal/domain/step"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type StepHandler struct {
	createHandler         *stepcmd.CreateStepHandler
	getByIDHandler        *querystep.GetStepByIDHandler
	listByWorkflowHandler *querystep.ListStepsByWorkflowHandler
	getWorkflowHandler    *queryworkflow.GetWorkflowByIDHandler
}

func NewStepHandler(
	createHandler *stepcmd.CreateStepHandler,
	getByIDHandler *querystep.GetStepByIDHandler,
	listByWorkflowHandler *querystep.ListStepsByWorkflowHandler,
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler,
) *StepHandler {
	return &StepHandler{
		createHandler:         createHandler,
		getByIDHandler:        getByIDHandler,
		listByWorkflowHandler: listByWorkflowHandler,
		getWorkflowHandler:    getWorkflowHandler,
	}
}

func (h *StepHandler) Create(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	var req dto.CreateStepRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Index = strings.TrimSpace(req.Index)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	endpointID, err := uuid.Parse(req.EndpointID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid endpoint id"})
	}

	s, err := h.createHandler.Handle(c.Context(), stepcmd.CreateStepCommand{
		WorkflowID:     workflowID,
		EndpointID:     endpointID,
		OrganizationID: orgID,
		Index:          req.Index,
		TreeIndex:      0,
		Position: domainstep.Position{
			X: req.Position.X,
			Y: req.Position.Y,
		},
	})
	if err != nil {
		switch err.Error() {
		case "workflow not found", "endpoint not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create step"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewStepDetailResponseFromEntity(*s))
}

func (h *StepHandler) GetByID(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get step"})
	}
	if view.OrganizationID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromView(*view))
}

func (h *StepHandler) ListByWorkflow(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}

	var listQuery dto.ListStepsQuery
	if err := c.Bind().Query(&listQuery); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	if err := validation.Struct(c, &listQuery); err != nil {
		return err
	}

	workflowID, err := uuid.Parse(listQuery.WorkflowID)
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

	views, err := h.listByWorkflowHandler.Handle(c.Context(), querystep.ListStepsByWorkflowQuery{
		WorkflowID: workflowID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list steps"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewStepListResponseFromViews(views))
}
