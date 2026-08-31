package handler

import (
	"errors"

	assertioncmd "go-api/internal/application/command/assertion"
	queryassertion "go-api/internal/application/query/assertion"
	querystep "go-api/internal/application/query/step"
	queryworkflow "go-api/internal/application/query/workflow"
	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AssertionHandler struct {
	createHandler      assertionCreateHandler
	updateHandler      assertionUpdateHandler
	deleteHandler      assertionDeleteHandler
	getByIDHandler     assertionGetByIDHandler
	listByStep         assertionListByStepHandler
	searchPaths        assertionSearchPathsHandler
	getStepHandler     assertionGetStepHandler
	getWorkflowHandler assertionGetWorkflowHandler
}

func NewAssertionHandler(
	createHandler assertionCreateHandler,
	updateHandler assertionUpdateHandler,
	deleteHandler assertionDeleteHandler,
	getByIDHandler assertionGetByIDHandler,
	listByStep assertionListByStepHandler,
	searchPaths assertionSearchPathsHandler,
	getStepHandler assertionGetStepHandler,
	getWorkflowHandler assertionGetWorkflowHandler,
) *AssertionHandler {
	return &AssertionHandler{
		createHandler:      createHandler,
		updateHandler:      updateHandler,
		deleteHandler:      deleteHandler,
		getByIDHandler:     getByIDHandler,
		listByStep:         listByStep,
		searchPaths:        searchPaths,
		getStepHandler:     getStepHandler,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *AssertionHandler) ensureWorkflow(c fiber.Ctx, workflowID, projectID uuid.UUID) (int, string) {
	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return fiber.StatusNotFound, "Workflow not found"
		}
		return fiber.StatusInternalServerError, "Failed to get workflow"
	}
	if workflow.ProjectID != projectID {
		return fiber.StatusNotFound, "Workflow not found"
	}
	return 0, ""
}

func expectedValueFromRequest(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (h *AssertionHandler) Create(c fiber.Ctx) error {
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
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	var req dto.CreateAssertionRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}

	source, err := domainassertion.ParseSource(req.Source)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion source"})
	}
	operator, err := domainassertion.ParseOperator(req.Operator)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion operator"})
	}

	assertion, err := h.createHandler.Handle(c.Context(), assertioncmd.CreateAssertionCommand{
		UserID:        user.ID,
		WorkflowID:    workflowID,
		ProjectID:     projectID,
		StepID:        stepID,
		Description:   req.Description,
		Source:        source,
		Path:          req.Path,
		Operator:      operator,
		ExpectedValue: expectedValueFromRequest(req.ExpectedValue),
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(presenter.NewAssertionResponseFromEntity(*assertion))
}

func (h *AssertionHandler) ListByStep(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	step, err := h.getStepHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: stepID})
	if err != nil || step.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	views, err := h.listByStep.Handle(c.Context(), queryassertion.ListAssertionsByStepQuery{StepID: stepID})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list assertions"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewAssertionListResponseFromViews(views))
}

func (h *AssertionHandler) SearchPaths(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	stepID, err := uuid.Parse(c.Params("stepId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
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

	paths, total, err := h.searchPaths.Handle(c.Context(), queryassertion.SearchAssertionPathsQuery{
		WorkflowID: workflowID,
		StepID:     stepID,
		Query:      query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to search assertion paths"})
	}
	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(paths, total, query))
}

func (h *AssertionHandler) GetByID(c fiber.Ctx) error {
	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}
	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryassertion.GetAssertionByIDQuery{
		ID:         id,
		WorkflowID: workflowID,
	})
	if err != nil {
		if errors.Is(err, domainassertion.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Assertion not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get assertion"})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewAssertionResponseFromView(*view))
}

func (h *AssertionHandler) Update(c fiber.Ctx) error {
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
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	var req dto.UpdateAssertionRequest
	if err := validation.BindBody(c, &req); err != nil {
		return err
	}

	source, err := domainassertion.ParseSource(req.Source)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion source"})
	}
	operator, err := domainassertion.ParseOperator(req.Operator)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion operator"})
	}

	assertion, err := h.updateHandler.Handle(c.Context(), assertioncmd.UpdateAssertionCommand{
		ID:            id,
		UserID:        user.ID,
		WorkflowID:    workflowID,
		ProjectID:     projectID,
		Description:   req.Description,
		Source:        source,
		Path:          req.Path,
		Operator:      operator,
		ExpectedValue: expectedValueFromRequest(req.ExpectedValue),
	})
	if err != nil {
		if errors.Is(err, domainassertion.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Assertion not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(presenter.NewAssertionResponseFromEntity(*assertion))
}

func (h *AssertionHandler) Delete(c fiber.Ctx) error {
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
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid assertion id"})
	}
	if code, msg := h.ensureWorkflow(c, workflowID, projectID); code != 0 {
		return c.Status(code).JSON(fiber.Map{"message": msg})
	}

	if err := h.deleteHandler.Handle(c.Context(), assertioncmd.DeleteAssertionCommand{
		ID:         id,
		UserID:     user.ID,
		WorkflowID: workflowID,
		ProjectID:  projectID,
	}); err != nil {
		if errors.Is(err, domainassertion.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Assertion not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete assertion"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}
