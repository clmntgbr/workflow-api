package handler

import (
	"errors"

	variablecmd "go-api/internal/application/command/variable"
	querystep "go-api/internal/application/query/step"
	queryvariable "go-api/internal/application/query/variable"
	queryworkflow "go-api/internal/application/query/workflow"
	"go-api/internal/domain/paginate"
	domainvariable "go-api/internal/domain/variable"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type VariableHandler struct {
	createHandler      *variablecmd.CreateVariableHandler
	updateHandler      *variablecmd.UpdateVariableHandler
	deleteHandler      *variablecmd.DeleteVariableHandler
	getByIDHandler     *queryvariable.GetVariableByIDHandler
	listByWorkflow     *queryvariable.ListVariablesByWorkflowHandler
	listAvailable      *queryvariable.ListAvailableVariablesHandler
	searchPaths        *queryvariable.SearchVariablePathsHandler
	getStepHandler     *querystep.GetStepByIDHandler
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler
}

func NewVariableHandler(
	createHandler *variablecmd.CreateVariableHandler,
	updateHandler *variablecmd.UpdateVariableHandler,
	deleteHandler *variablecmd.DeleteVariableHandler,
	getByIDHandler *queryvariable.GetVariableByIDHandler,
	listByWorkflow *queryvariable.ListVariablesByWorkflowHandler,
	listAvailable *queryvariable.ListAvailableVariablesHandler,
	searchPaths *queryvariable.SearchVariablePathsHandler,
	getStepHandler *querystep.GetStepByIDHandler,
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler,
) *VariableHandler {
	return &VariableHandler{
		createHandler:      createHandler,
		updateHandler:      updateHandler,
		deleteHandler:      deleteHandler,
		getByIDHandler:     getByIDHandler,
		listByWorkflow:     listByWorkflow,
		listAvailable:      listAvailable,
		searchPaths:        searchPaths,
		getStepHandler:     getStepHandler,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *VariableHandler) ensureWorkflow(c fiber.Ctx, workflowID, orgID uuid.UUID) (int, string) {
	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return fiber.StatusNotFound, "Workflow not found"
		}
		return fiber.StatusInternalServerError, "Failed to get workflow"
	}
	if workflow.OrganizationID != orgID {
		return fiber.StatusNotFound, "Workflow not found"
	}
	return 0, ""
}

func (h *VariableHandler) Create(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	var req dto.CreateVariableRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}
	stepID, err := uuid.Parse(req.StepID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	variable, err := h.createHandler.Handle(c.Context(), variablecmd.CreateVariableCommand{
		WorkflowID:     workflowID,
		OrganizationID: orgID,
		StepID:         stepID,
		Name:           req.Name,
		Key:            req.Key,
		Description:    req.Description,
		Path:           req.Path,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainvariable.ErrDuplicateKey):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Variable key already exists"})
		case err.Error() == "step not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	}
	return c.Status(fiber.StatusCreated).JSON(presenter.NewVariableResponseFromEntity(*variable))
}

func (h *VariableHandler) List(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	views, err := h.listByWorkflow.Handle(c.Context(), queryvariable.ListVariablesByWorkflowQuery{WorkflowID: workflowID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list variables"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewVariableListResponseFromViews(views))
}

func (h *VariableHandler) ListAvailable(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	step, err := h.getStepHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: stepID})
	if err != nil || step.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	views, err := h.listAvailable.Handle(c.Context(), queryvariable.ListAvailableVariablesQuery{
		WorkflowID: workflowID,
		StepID:     stepID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list available variables"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewVariableListResponseFromViews(views))
}

func (h *VariableHandler) SearchPaths(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveOrganizationID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active organization is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	step, err := h.getStepHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: stepID})
	if err != nil || step.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	query.Normalize()

	paths, total, err := h.searchPaths.Handle(c.Context(), queryvariable.SearchVariablePathsQuery{
		WorkflowID: workflowID,
		StepID:     stepID,
		Query:      query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to search variable paths"})
	}
	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(paths, total, query))
}

func (h *VariableHandler) GetByID(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid variable id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryvariable.GetVariableByIDQuery{
		ID:         id,
		WorkflowID: workflowID,
	})
	if err != nil {
		if errors.Is(err, domainvariable.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Variable not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get variable"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewVariableResponseFromView(*view))
}

func (h *VariableHandler) Update(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid variable id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	var req dto.UpdateVariableRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}

	variable, err := h.updateHandler.Handle(c.Context(), variablecmd.UpdateVariableCommand{
		ID:             id,
		WorkflowID:     workflowID,
		OrganizationID: orgID,
		Name:           req.Name,
		Key:            req.Key,
		Description:    req.Description,
		Path:           req.Path,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainvariable.ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Variable not found"})
		case errors.Is(err, domainvariable.ErrDuplicateKey):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Variable key already exists"})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewVariableResponseFromEntity(*variable))
}

func (h *VariableHandler) Delete(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid variable id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, orgID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	if err := h.deleteHandler.Handle(c.Context(), variablecmd.DeleteVariableCommand{
		ID:         id,
		WorkflowID: workflowID,
	}); err != nil {
		if errors.Is(err, domainvariable.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Variable not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete variable"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
