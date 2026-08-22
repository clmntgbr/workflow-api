package handler

import (
	"errors"
	"strings"
	"time"

	workflowruncmd "go-api/internal/application/command/workflowrun"
	queryinsight "go-api/internal/application/query/insight"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	domaininsight "go-api/internal/domain/insight"
	"go-api/internal/domain/paginate"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
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
	cancelHandler      *workflowruncmd.CancelWorkflowRunHandler
	getByIDHandler     *queryworkflowrun.GetWorkflowRunByIDHandler
	analyticsHandler   *queryworkflowrun.GetWorkflowRunAnalyticsHandler
	listByWorkflow     *queryworkflowrun.ListWorkflowRunsByWorkflowHandler
	listByProject *queryworkflowrun.ListWorkflowRunsByProjectHandler
	listStepRuns       *querysteprun.ListStepRunsByWorkflowRunHandler
	listStepRunsByIDs  *querysteprun.ListStepRunsByWorkflowRunIDsHandler
	listInsightsByIDs  *queryinsight.ListInsightsByStepRunIDsHandler
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler
	listStepsByWorkflow *querystep.ListStepsByWorkflowHandler
}

func NewWorkflowRunHandler(
	startHandler *workflowruncmd.StartWorkflowRunHandler,
	cancelHandler *workflowruncmd.CancelWorkflowRunHandler,
	getByIDHandler *queryworkflowrun.GetWorkflowRunByIDHandler,
	analyticsHandler *queryworkflowrun.GetWorkflowRunAnalyticsHandler,
	listByWorkflow *queryworkflowrun.ListWorkflowRunsByWorkflowHandler,
	listByProject *queryworkflowrun.ListWorkflowRunsByProjectHandler,
	listStepRuns *querysteprun.ListStepRunsByWorkflowRunHandler,
	listStepRunsByIDs *querysteprun.ListStepRunsByWorkflowRunIDsHandler,
	listInsightsByIDs *queryinsight.ListInsightsByStepRunIDsHandler,
	getWorkflowHandler *queryworkflow.GetWorkflowByIDHandler,
	listStepsByWorkflow *querystep.ListStepsByWorkflowHandler,
) *WorkflowRunHandler {
	return &WorkflowRunHandler{
		startHandler:        startHandler,
		cancelHandler:       cancelHandler,
		getByIDHandler:      getByIDHandler,
		analyticsHandler:    analyticsHandler,
		listByWorkflow:      listByWorkflow,
		listByProject:  listByProject,
		listStepRuns:        listStepRuns,
		listStepRunsByIDs:   listStepRunsByIDs,
		listInsightsByIDs:   listInsightsByIDs,
		getWorkflowHandler:  getWorkflowHandler,
		listStepsByWorkflow: listStepsByWorkflow,
	}
}

func (h *WorkflowRunHandler) Analytics(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	var query dto.WorkflowRunAnalyticsQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
			"errors":  err.Error(),
		})
	}
	if err := validation.Struct(c, &query); err != nil {
		return err
	}

	var workflowID *uuid.UUID
	if query.WorkflowID != "" {
		parsed, err := uuid.Parse(query.WorkflowID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
		}
		workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: parsed})
		if err != nil {
			if err.Error() == "workflow not found" {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
		}
		if workflow.ProjectID != orgID {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		workflowID = &parsed
	}

	from, err := parseRFC3339Nullable(query.From)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid from date (expected RFC3339)"})
	}
	to, err := parseRFC3339Nullable(query.To)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid to date (expected RFC3339)"})
	}

	stats, err := h.analyticsHandler.Handle(c.Context(), queryworkflowrun.GetWorkflowRunAnalyticsQuery{
		ProjectID: orgID,
		WorkflowID:     workflowID,
		From:           from,
		To:             to,
	})
	if err != nil {
		if strings.Contains(err.Error(), "from must be before to") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "from must be before to"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to compute workflow run analytics"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowRunAnalyticsResponse(*stats))
}

func parseRFC3339Nullable(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (h *WorkflowRunHandler) StartWorkflow(c fiber.Ctx) error {
	return h.startWorkflowRun(c, c.Params("id"))
}

func (h *WorkflowRunHandler) StopWorkflow(c fiber.Ctx) error {
	if _, err := httpctx.GetUser(c); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("id"))
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

	run, err := h.cancelHandler.Handle(c.Context(), workflowruncmd.CancelWorkflowRunCommand{
		WorkflowID:     workflowID,
		ProjectID: orgID,
	})
	if err != nil {
		if errors.Is(err, domainworkflowrun.ErrWorkflowNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if errors.Is(err, domainworkflowrun.ErrNoRunInProgress) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "NO_RUN_IN_PROGRESS",
				"message": "No workflow run is in progress",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to stop workflow run"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowRunDetailResponseFromEntity(*run, orgID))
}

func (h *WorkflowRunHandler) startWorkflowRun(
	c fiber.Ctx,
	workflowIDParam string,
) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(workflowIDParam)
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

	var req dto.StartWorkflowRunRequest
	if len(c.Body()) > 0 {
		if err := validation.BindBody(c, &req); err != nil {
			return err
		}
	}

	userID := user.ID
	run, err := h.startHandler.Handle(c.Context(), workflowruncmd.StartWorkflowRunCommand{
		WorkflowID:        workflowID,
		TriggeredBy:       domainworkflowrun.TriggeredByAPI,
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
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to start workflow run"})
	}

	return c.Status(fiber.StatusCreated).JSON(presenter.NewWorkflowRunDetailResponseFromEntity(*run, orgID))
}

func (h *WorkflowRunHandler) ListByWorkflow(c fiber.Ctx) error {
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

	ids := make([]uuid.UUID, 0, len(views))
	for _, view := range views {
		ids = append(ids, view.ID)
	}

	stepRuns, err := h.listStepRunsByIDs.Handle(c.Context(), querysteprun.ListStepRunsByWorkflowRunIDsQuery{
		WorkflowRunIDs: ids,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list step runs"})
	}

	stepRunsByWorkflowRunID := make(map[uuid.UUID][]domainsteprun.StepRunView, len(views))
	for _, stepRun := range stepRuns {
		stepRunsByWorkflowRunID[stepRun.WorkflowRunID] = append(
			stepRunsByWorkflowRunID[stepRun.WorkflowRunID],
			stepRun,
		)
	}

	insightsByStepRunID, err := h.loadInsightsByStepRuns(c, stepRuns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list insights"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewWorkflowRunListWithStepRunsFromViews(views, stepRunsByWorkflowRunID, insightsByStepRunID),
		int(total),
		query,
	))
}

func (h *WorkflowRunHandler) GetByID(c fiber.Ctx) error {
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow run id"})
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

	insightsByStepRunID, err := h.loadInsightsByStepRuns(c, stepRuns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list insights"})
	}

	stepsByID, err := h.loadStepsByWorkflow(c, view.WorkflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list steps"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowRunDetailResponseFromViewWithRelations(
		*view,
		workflow,
		stepRuns,
		stepsByID,
		insightsByStepRunID,
	))
}

func (h *WorkflowRunHandler) List(c fiber.Ctx) error {
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
		query.SortBy = "workflow_runs.created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByDesc
	}

	views, total, err := h.listByProject.Handle(c.Context(), queryworkflowrun.ListWorkflowRunsByProjectQuery{
		ProjectID: orgID,
		Query:          query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list workflow runs"})
	}

	ids := make([]uuid.UUID, 0, len(views))
	for _, view := range views {
		ids = append(ids, view.ID)
	}

	stepRuns, err := h.listStepRunsByIDs.Handle(c.Context(), querysteprun.ListStepRunsByWorkflowRunIDsQuery{
		WorkflowRunIDs: ids,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list step runs"})
	}

	stepRunsByWorkflowRunID := make(map[uuid.UUID][]domainsteprun.StepRunView, len(views))
	for _, stepRun := range stepRuns {
		stepRunsByWorkflowRunID[stepRun.WorkflowRunID] = append(
			stepRunsByWorkflowRunID[stepRun.WorkflowRunID],
			stepRun,
		)
	}

	insightsByStepRunID, err := h.loadInsightsByStepRuns(c, stepRuns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list insights"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewWorkflowRunListWithStepRunsFromViews(views, stepRunsByWorkflowRunID, insightsByStepRunID),
		int(total),
		query,
	))
}

func (h *WorkflowRunHandler) Get(c fiber.Ctx) error {
	orgID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow run id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryworkflowrun.GetWorkflowRunByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "workflow run not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow run not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow run"})
	}
	if view.ProjectID != orgID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow run not found"})
	}

	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: view.WorkflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}

	stepRuns, err := h.listStepRuns.Handle(c.Context(), querysteprun.ListStepRunsByWorkflowRunQuery{
		WorkflowRunID: view.ID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list step runs"})
	}

	insightsByStepRunID, err := h.loadInsightsByStepRuns(c, stepRuns)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list insights"})
	}

	stepsByID, err := h.loadStepsByWorkflow(c, view.WorkflowID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list steps"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewWorkflowRunDetailResponseFromViewWithRelations(
		*view,
		workflow,
		stepRuns,
		stepsByID,
		insightsByStepRunID,
	))
}

func (h *WorkflowRunHandler) loadInsightsByStepRuns(
	c fiber.Ctx,
	stepRuns []domainsteprun.StepRunView,
) (map[uuid.UUID][]domaininsight.InsightView, error) {
	ids := make([]uuid.UUID, 0, len(stepRuns))
	for _, stepRun := range stepRuns {
		ids = append(ids, stepRun.ID)
	}

	insights, err := h.listInsightsByIDs.Handle(c.Context(), queryinsight.ListInsightsByStepRunIDsQuery{
		StepRunIDs: ids,
	})
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID][]domaininsight.InsightView, len(ids))
	for _, insight := range insights {
		out[insight.StepRunID] = append(out[insight.StepRunID], insight)
	}
	return out, nil
}

func (h *WorkflowRunHandler) loadStepsByWorkflow(
	c fiber.Ctx,
	workflowID uuid.UUID,
) (map[uuid.UUID]domainstep.StepView, error) {
	steps, err := h.listStepsByWorkflow.Handle(c.Context(), querystep.ListStepsByWorkflowQuery{
		WorkflowID: workflowID,
	})
	if err != nil {
		return nil, err
	}

	out := make(map[uuid.UUID]domainstep.StepView, len(steps))
	for _, step := range steps {
		out[step.ID] = step
	}
	return out, nil
}
