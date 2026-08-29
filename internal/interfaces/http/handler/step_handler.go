package handler

import (
	"errors"
	"strings"

	stepcmd "go-api/internal/application/command/step"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type StepHandler struct {
	createHandler              *stepcmd.CreateStepHandler
	createDelayHandler         *stepcmd.CreateDelayStepHandler
	createConditionHandler     *stepcmd.CreateConditionStepHandler
	updateHandler              *stepcmd.UpdateStepHandler
	updateDelayHandler         *stepcmd.UpdateDelayStepHandler
	updateConditionHandler     *stepcmd.UpdateConditionStepHandler
	updatePositionHandler      *stepcmd.UpdateStepPositionHandler
	deleteHandler              *stepcmd.DeleteStepHandler
	getByIDHandler             *querystep.GetStepByIDHandler
	listByWorkflowHandler      *querystep.ListStepsByWorkflowHandler
	latestStepRunStatusHandler *querysteprun.GetLatestStepRunStatusesByStepIDsHandler
	getWorkflowHandler         *queryworkflow.GetWorkflowByIDHandler
}

func NewStepHandler(
	createHandler *stepcmd.CreateStepHandler,
	createDelayHandler *stepcmd.CreateDelayStepHandler,
	createConditionHandler *stepcmd.CreateConditionStepHandler,
	updateHandler *stepcmd.UpdateStepHandler,
	updateDelayHandler *stepcmd.UpdateDelayStepHandler,
	updateConditionHandler *stepcmd.UpdateConditionStepHandler,
	updatePositionHandler *stepcmd.UpdateStepPositionHandler,
	deleteHandler *stepcmd.DeleteStepHandler,
	getByIDHandler *querystep.GetStepByIDHandler,
	listByWorkflowHandler *querystep.ListStepsByWorkflowHandler,
	latestStepRunStatusHandler *querysteprun.GetLatestStepRunStatusesByStepIDsHandler,
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler,
) *StepHandler {
	return &StepHandler{
		createHandler:              createHandler,
		createDelayHandler:         createDelayHandler,
		createConditionHandler:     createConditionHandler,
		updateHandler:              updateHandler,
		updateDelayHandler:         updateDelayHandler,
		updateConditionHandler:     updateConditionHandler,
		updatePositionHandler:      updatePositionHandler,
		deleteHandler:              deleteHandler,
		getByIDHandler:             getByIDHandler,
		listByWorkflowHandler:      listByWorkflowHandler,
		latestStepRunStatusHandler: latestStepRunStatusHandler,
		getWorkflowHandler:         getWorkflowHandler,
	}
}

func (h *StepHandler) Create(c fiber.Ctx) error {
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

	var req dto.CreateStepRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	stepType := req.Type
	if stepType == "" {
		stepType = string(domainstep.TypeHTTP)
	}

	if stepType == string(domainstep.TypeDelay) {
		if req.DelayDurationSeconds == nil || *req.DelayDurationSeconds <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "delayDurationSeconds is required for delay steps"})
		}
		s, err := h.createDelayHandler.Handle(c.Context(), stepcmd.CreateDelayStepCommand{
			UserID:               user.ID,
			WorkflowID:           workflowID,
			ProjectID:            orgID,
			Name:                 req.Name,
			DelayDurationSeconds: *req.DelayDurationSeconds,
			Position: domainstep.Position{
				X: req.Position.X,
				Y: req.Position.Y,
			},
		})
		if err != nil {
			if handled, resp := respondQuotaError(c, err); handled {
				return resp
			}
			switch {
			case errors.Is(err, domainstep.ErrInvalidStepTypeConfig):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
			case err.Error() == "workflow not found":
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create step"})
			}
		}
		return c.Status(fiber.StatusCreated).JSON(presenter.NewStepDetailResponseFromEntity(*s))
	}

	if stepType == string(domainstep.TypeCondition) {
		if req.Expression == nil || strings.TrimSpace(*req.Expression) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "expression is required for condition steps"})
		}
		s, err := h.createConditionHandler.Handle(c.Context(), stepcmd.CreateConditionStepCommand{
			UserID:     user.ID,
			WorkflowID: workflowID,
			ProjectID:  orgID,
			Name:       req.Name,
			Expression: strings.TrimSpace(*req.Expression),
			Position: domainstep.Position{
				X: req.Position.X,
				Y: req.Position.Y,
			},
		})
		if err != nil {
			if handled, resp := respondQuotaError(c, err); handled {
				return resp
			}
			switch {
			case errors.Is(err, domainstep.ErrInvalidStepTypeConfig):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
			case err.Error() == "workflow not found":
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": err.Error()})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create step"})
			}
		}
		return c.Status(fiber.StatusCreated).JSON(presenter.NewStepDetailResponseFromEntity(*s))
	}

	if req.EndpointID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "endpointId is required for HTTP steps"})
	}

	endpointID, err := uuid.Parse(req.EndpointID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid endpoint id"})
	}

	s, err := h.createHandler.Handle(c.Context(), stepcmd.CreateStepCommand{
		UserID:         user.ID,
		WorkflowID:     workflowID,
		EndpointID:     endpointID,
		ProjectID: orgID,
		Position: domainstep.Position{
			X: req.Position.X,
			Y: req.Position.Y,
		},
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get step"})
	}
	if view.ProjectID != orgID || view.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromView(*view))
}

func (h *StepHandler) ListByWorkflow(c fiber.Ctx) error {
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

	views, err := h.listByWorkflowHandler.Handle(c.Context(), querystep.ListStepsByWorkflowQuery{
		WorkflowID: workflowID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list steps"})
	}

	stepIDs := make([]uuid.UUID, 0, len(views))
	for _, view := range views {
		stepIDs = append(stepIDs, view.ID)
	}

	statuses, err := h.latestStepRunStatusHandler.Handle(c.Context(), querysteprun.GetLatestStepRunStatusesByStepIDsQuery{
		WorkflowID: workflowID,
		StepIDs:    stepIDs,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to load step run statuses"})
	}

	lastRunStatusByStepID := make(map[uuid.UUID]string, len(statuses))
	for stepID, status := range statuses {
		lastRunStatusByStepID[stepID] = string(status)
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewStepListResponseFromViewsWithLastRunStatus(views, lastRunStatusByStepID),
	)
}

func (h *StepHandler) Update(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	existing, err := h.getByIDHandler.Handle(c.Context(), querystep.GetStepByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get step"})
	}
	if existing.ProjectID != orgID || existing.WorkflowID != workflowID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
	}

	if existing.Type == domainstep.TypeDelay {
		var req dto.UpdateDelayStepRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Description = strings.TrimSpace(req.Description)
		if err := validation.Struct(c, &req); err != nil {
			return err
		}
		s, err := h.updateDelayHandler.Handle(c.Context(), stepcmd.UpdateDelayStepCommand{
			ID:                   id,
			UserID:               user.ID,
			WorkflowID:           workflowID,
			ProjectID:            orgID,
			Name:                 req.Name,
			Description:          req.Description,
			DelayDurationSeconds: req.DelayDurationSeconds,
		})
		if err != nil {
			switch {
			case errors.Is(err, domainstep.ErrInvalidStepTypeConfig):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
			case err.Error() == "step not found":
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update step"})
			}
		}
		return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromEntity(*s))
	}

	if existing.Type == domainstep.TypeCondition {
		var req dto.UpdateConditionStepRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
		}
		req.Name = strings.TrimSpace(req.Name)
		req.Description = strings.TrimSpace(req.Description)
		req.Expression = strings.TrimSpace(req.Expression)
		if err := validation.Struct(c, &req); err != nil {
			return err
		}
		s, err := h.updateConditionHandler.Handle(c.Context(), stepcmd.UpdateConditionStepCommand{
			ID:          id,
			UserID:      user.ID,
			WorkflowID:  workflowID,
			ProjectID:   orgID,
			Name:        req.Name,
			Description: req.Description,
			Expression:  req.Expression,
		})
		if err != nil {
			switch {
			case errors.Is(err, domainstep.ErrInvalidStepTypeConfig):
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
			case err.Error() == "step not found":
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update step"})
			}
		}
		return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromEntity(*s))
	}

	var req dto.UpdateStepRequest
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

	s, err := h.updateHandler.Handle(c.Context(), stepcmd.UpdateStepCommand{
		ID:             id,
		UserID:         user.ID,
		WorkflowID:     workflowID,
		ProjectID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Method:         req.Method,
		Headers:        headers,
		Query:          queryParams,
		Body:           body,
		Timeout:        *req.Timeout,
		RetryOnFailure: *req.RetryOnFailure,
		RetryCount:     *req.RetryCount,
		RetryDelay:     *req.RetryDelay,
	})
	if err != nil {
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update step"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromEntity(*s))
}

func (h *StepHandler) UpdatePosition(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	var req dto.UpdateStepPositionRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	s, err := h.updatePositionHandler.Handle(c.Context(), stepcmd.UpdateStepPositionCommand{
		ID:         id,
		UserID:     user.ID,
		ProjectID:  orgID,
		WorkflowID: workflowID,
		Position: domainstep.Position{
			X: req.Position.X,
			Y: req.Position.Y,
		},
	})
	if err != nil {
		switch err.Error() {
		case "step not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update step"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewStepDetailResponseFromEntity(*s))
}

func (h *StepHandler) Delete(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid step id"})
	}

	if err := h.deleteHandler.Handle(c.Context(), stepcmd.DeleteStepCommand{
		ID:         id,
		UserID:     user.ID,
		WorkflowID: workflowID,
		ProjectID:  orgID,
	}); err != nil {
		if err.Error() == "step not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Step not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete step"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
