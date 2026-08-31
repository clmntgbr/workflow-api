package handler

import (
	"errors"

	conncmd "go-api/internal/application/command/connection"
	queryconn "go-api/internal/application/query/connection"
	queryworkflow "go-api/internal/application/query/workflow"
	domainconnection "go-api/internal/domain/connection"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ConnectionHandler struct {
	createHandler      connectionCreateHandler
	deleteHandler      connectionDeleteHandler
	listByWorkflow     connectionListByWorkflowHandler
	getWorkflowHandler connectionGetWorkflowHandler
}

func NewConnectionHandler(
	createHandler connectionCreateHandler,
	deleteHandler connectionDeleteHandler,
	listByWorkflow connectionListByWorkflowHandler,
	getWorkflowHandler connectionGetWorkflowHandler,
) *ConnectionHandler {
	return &ConnectionHandler{
		createHandler:      createHandler,
		deleteHandler:      deleteHandler,
		listByWorkflow:     listByWorkflow,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *ConnectionHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	var req dto.CreateConnectionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	sourceStepID, err := uuid.Parse(req.SourceStepID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid source step id"})
	}
	targetStepID, err := uuid.Parse(req.TargetStepID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid target step id"})
	}

	var branch *domainconnection.ConditionBranch
	if req.Branch != nil {
		parsed, err := domainconnection.ParseConditionBranch(*req.Branch)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
		branch = &parsed
	}

	conn, err := h.createHandler.Handle(c.Context(), conncmd.CreateConnectionCommand{
		UserID:       user.ID,
		WorkflowID:   workflowID,
		ProjectID:    orgID,
		SourceStepID: sourceStepID,
		TargetStepID: targetStepID,
		Branch:       branch,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainconnection.ErrInvalidBranch),
			errors.Is(err, domainconnection.ErrConditionRequiresBranch),
			errors.Is(err, domainconnection.ErrNonConditionBranchForbidden),
			errors.Is(err, domainconnection.ErrConditionOutgoingCount),
			errors.Is(err, domainconnection.ErrConditionalTargetMultipleParents):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		case err.Error() == "source step not found", err.Error() == "target step not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
		case err.Error() == "source and target steps must be different":
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create connection"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewConnectionDetailResponseFromEntity(*conn))
}

func (h *ConnectionHandler) Delete(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid connection id"})
	}

	if err := h.deleteHandler.Handle(c.Context(), conncmd.DeleteConnectionCommand{
		ID:         id,
		UserID:     user.ID,
		WorkflowID: workflowID,
		ProjectID:  orgID,
	}); err != nil {
		if err.Error() == "connection not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Connection not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete connection"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ConnectionHandler) ListByWorkflow(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
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
	if workflow.ProjectID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	views, err := h.listByWorkflow.Handle(c.Context(), queryconn.ListConnectionsByWorkflowQuery{
		WorkflowID: workflowID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list connections"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewConnectionListResponseFromViews(views))
}
