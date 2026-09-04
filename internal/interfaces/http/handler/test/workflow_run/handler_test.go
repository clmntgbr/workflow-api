package workflowruntest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	queryinsight "go-api/internal/application/query/insight"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	domaininsight "go-api/internal/domain/insight"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var (
	otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")
	testStepRunID  = uuid.MustParse("01960000-0000-7000-8000-00000000000c")
)

type mockStartWorkflowRunHandler struct {
	called bool
	cmd    workflowruncmd.StartWorkflowRunCommand
	result *domainworkflowrun.WorkflowRun
	err    error
}

func (m *mockStartWorkflowRunHandler) Handle(
	_ context.Context,
	cmd workflowruncmd.StartWorkflowRunCommand,
) (*domainworkflowrun.WorkflowRun, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockCancelWorkflowRunHandler struct {
	called bool
	cmd    workflowruncmd.CancelWorkflowRunCommand
	result *domainworkflowrun.WorkflowRun
	err    error
}

func (m *mockCancelWorkflowRunHandler) Handle(
	_ context.Context,
	cmd workflowruncmd.CancelWorkflowRunCommand,
) (*domainworkflowrun.WorkflowRun, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockGetWorkflowRunByIDHandler struct {
	called bool
	query  queryworkflowrun.GetWorkflowRunByIDQuery
	view   *domainworkflowrun.WorkflowRunView
	err    error
}

func (m *mockGetWorkflowRunByIDHandler) Handle(
	_ context.Context,
	q queryworkflowrun.GetWorkflowRunByIDQuery,
) (*domainworkflowrun.WorkflowRunView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockWorkflowRunAnalyticsHandler struct {
	called bool
	query  queryworkflowrun.GetWorkflowRunAnalyticsQuery
	stats  *domainworkflowrun.WorkflowRunAnalytics
	err    error
}

func (m *mockWorkflowRunAnalyticsHandler) Handle(
	_ context.Context,
	q queryworkflowrun.GetWorkflowRunAnalyticsQuery,
) (*domainworkflowrun.WorkflowRunAnalytics, error) {
	m.called = true
	m.query = q
	return m.stats, m.err
}

type mockListWorkflowRunsByWorkflowHandler struct {
	called bool
	query  queryworkflowrun.ListWorkflowRunsByWorkflowQuery
	views  []domainworkflowrun.WorkflowRunView
	total  int64
	err    error
}

func (m *mockListWorkflowRunsByWorkflowHandler) Handle(
	_ context.Context,
	q queryworkflowrun.ListWorkflowRunsByWorkflowQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockListStepRunsByWorkflowRunHandler struct {
	called bool
	query  querysteprun.ListStepRunsByWorkflowRunQuery
	views  []domainsteprun.StepRunView
	err    error
}

func (m *mockListStepRunsByWorkflowRunHandler) Handle(
	_ context.Context,
	q querysteprun.ListStepRunsByWorkflowRunQuery,
) ([]domainsteprun.StepRunView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockListStepRunsByWorkflowRunIDsHandler struct {
	called bool
	query  querysteprun.ListStepRunsByWorkflowRunIDsQuery
	views  []domainsteprun.StepRunView
	err    error
}

func (m *mockListStepRunsByWorkflowRunIDsHandler) Handle(
	_ context.Context,
	q querysteprun.ListStepRunsByWorkflowRunIDsQuery,
) ([]domainsteprun.StepRunView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockListInsightsByStepRunIDsHandler struct {
	called bool
	query  queryinsight.ListInsightsByStepRunIDsQuery
	views  []domaininsight.InsightView
	err    error
}

func (m *mockListInsightsByStepRunIDsHandler) Handle(
	_ context.Context,
	q queryinsight.ListInsightsByStepRunIDsQuery,
) ([]domaininsight.InsightView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockGetWorkflowByIDHandler struct {
	calls int
	views []*domainworkflow.WorkflowView
	errs  []error
}

func (m *mockGetWorkflowByIDHandler) Handle(
	_ context.Context,
	_ queryworkflow.GetWorkflowByIDQuery,
) (*domainworkflow.WorkflowView, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get workflow call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	return m.views[idx], nil
}

type mockListStepsByWorkflowHandler struct {
	called bool
	query  querystep.ListStepsByWorkflowQuery
	views  []domainstep.StepView
	err    error
}

func (m *mockListStepsByWorkflowHandler) Handle(
	_ context.Context,
	q querystep.ListStepsByWorkflowQuery,
) ([]domainstep.StepView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockInsightsAllowedChecker struct {
	called    bool
	userID    uuid.UUID
	projectID uuid.UUID
	allowed   bool
	err       error
}

func (m *mockInsightsAllowedChecker) InsightsAllowed(
	_ context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
) (bool, error) {
	m.called = true
	m.userID = userID
	m.projectID = projectID
	if m.err != nil {
		return false, m.err
	}
	return m.allowed, nil
}

func newWorkflowRunHandler(
	start *mockStartWorkflowRunHandler,
	cancel *mockCancelWorkflowRunHandler,
	getByID *mockGetWorkflowRunByIDHandler,
	analytics *mockWorkflowRunAnalyticsHandler,
	list *mockListWorkflowRunsByWorkflowHandler,
	listStepRuns *mockListStepRunsByWorkflowRunHandler,
	listStepRunsByIDs *mockListStepRunsByWorkflowRunIDsHandler,
	listInsights *mockListInsightsByStepRunIDsHandler,
	getWorkflow *mockGetWorkflowByIDHandler,
	listSteps *mockListStepsByWorkflowHandler,
	insightsAllowed *mockInsightsAllowedChecker,
) *handler.WorkflowRunHandler {
	if start == nil {
		start = &mockStartWorkflowRunHandler{}
	}
	if cancel == nil {
		cancel = &mockCancelWorkflowRunHandler{}
	}
	if getByID == nil {
		getByID = &mockGetWorkflowRunByIDHandler{}
	}
	if analytics == nil {
		analytics = &mockWorkflowRunAnalyticsHandler{}
	}
	if list == nil {
		list = &mockListWorkflowRunsByWorkflowHandler{}
	}
	if listStepRuns == nil {
		listStepRuns = &mockListStepRunsByWorkflowRunHandler{}
	}
	if listStepRunsByIDs == nil {
		listStepRunsByIDs = &mockListStepRunsByWorkflowRunIDsHandler{}
	}
	if listInsights == nil {
		listInsights = &mockListInsightsByStepRunIDsHandler{}
	}
	if getWorkflow == nil {
		getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	if listSteps == nil {
		listSteps = &mockListStepsByWorkflowHandler{}
	}
	if insightsAllowed == nil {
		insightsAllowed = &mockInsightsAllowedChecker{allowed: true}
	}
	return handler.NewWorkflowRunHandler(
		start, cancel, getByID, analytics, list,
		listStepRuns, listStepRunsByIDs, listInsights,
		getWorkflow, listSteps, insightsAllowed,
	)
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func sampleWorkflowView() *domainworkflow.WorkflowView {
	return &domainworkflow.WorkflowView{
		ID:        testutil.TestWorkflowID,
		Name:      "Order Flow",
		ProjectID: testutil.TestProjectID,
		Status:    domainworkflow.StatusActive,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleWorkflowRunEntity() *domainworkflowrun.WorkflowRun {
	started := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	return &domainworkflowrun.WorkflowRun{
		ID:          testutil.TestWorkflowRunID,
		WorkflowID:  testutil.TestWorkflowID,
		Status:      domainworkflowrun.StatusRunning,
		TriggeredBy: domainworkflowrun.TriggeredByAPI,
		StartedAt:   &started,
		CreatedAt:   time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
	}
}

func sampleWorkflowRunView() *domainworkflowrun.WorkflowRunView {
	started := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	return &domainworkflowrun.WorkflowRunView{
		ID:          testutil.TestWorkflowRunID,
		WorkflowID:  testutil.TestWorkflowID,
		ProjectID:   testutil.TestProjectID,
		Status:      domainworkflowrun.StatusRunning,
		TriggeredBy: domainworkflowrun.TriggeredByAPI,
		StartedAt:   &started,
		CreatedAt:   time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
	}
}

func sampleStepRunView() domainsteprun.StepRunView {
	return domainsteprun.StepRunView{
		ID:            testStepRunID,
		WorkflowRunID: testutil.TestWorkflowRunID,
		StepID:        testutil.TestStepID,
		WorkflowID:    testutil.TestWorkflowID,
		ProjectID:     testutil.TestProjectID,
		StepType:      domainstep.TypeHTTP,
		Name:          "Fetch order",
		Status:        domainsteprun.StatusSuccess,
		CreatedAt:     time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 1, 2, 10, 1, 0, 0, time.UTC),
	}
}

func sampleStepView() domainstep.StepView {
	return domainstep.StepView{
		ID:         testutil.TestStepID,
		WorkflowID: testutil.TestWorkflowID,
		ProjectID:  testutil.TestProjectID,
		Type:       domainstep.TypeHTTP,
		Name:       "Fetch order",
		CreatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleInsightView() domaininsight.InsightView {
	statusCode := 200
	return domaininsight.InsightView{
		ID:            uuid.New(),
		StepRunID:     testStepRunID,
		StatusCode:    &statusCode,
		AttemptNumber: 1,
		TotalAttempts: 1,
		CreatedAt:     time.Date(2026, 1, 2, 10, 1, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 1, 2, 10, 1, 0, 0, time.UTC),
	}
}

func workflowGetter(view *domainworkflow.WorkflowView, err error) *mockGetWorkflowByIDHandler {
	return &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{err},
	}
}

func mustJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	req, err := testutil.JSONRequest(method, path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	return req
}

func TestWorkflowRunHandler_StartWorkflow_Success(t *testing.T) {
	start := &mockStartWorkflowRunHandler{result: sampleWorkflowRunEntity()}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !start.called {
		t.Fatal("expected start handler to be called")
	}
	if start.cmd.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", start.cmd.WorkflowID)
	}
	if start.cmd.TriggeredBy != domainworkflowrun.TriggeredByAPI {
		t.Fatalf("triggered by: got %s", start.cmd.TriggeredBy)
	}

	var out presenter.WorkflowRunDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestWorkflowRunID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestWorkflowRunHandler_StartWorkflow_WithContext(t *testing.T) {
	start := &mockStartWorkflowRunHandler{result: sampleWorkflowRunEntity()}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	body := map[string]any{"context": map[string]any{"orderId": "123"}}
	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if start.cmd.Context["orderId"] != "123" {
		t.Fatalf("context: got %v", start.cmd.Context)
	}
}

func TestWorkflowRunHandler_StartWorkflow_Unauthorized(t *testing.T) {
	start := &mockStartWorkflowRunHandler{}
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if start.called {
		t.Fatal("start handler must not be called without user")
	}
}

func TestWorkflowRunHandler_StartWorkflow_MissingActiveProject(t *testing.T) {
	start := &mockStartWorkflowRunHandler{}
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", testutil.WithUserWithoutProject(testutil.TestUserID), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if start.called {
		t.Fatal("start handler must not be called without active project")
	}
}

func TestWorkflowRunHandler_StartWorkflow_InvalidWorkflowID(t *testing.T) {
	start := &mockStartWorkflowRunHandler{}
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/not-a-uuid/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if start.called {
		t.Fatal("start handler must not be called with invalid workflow id")
	}
}

func TestWorkflowRunHandler_StartWorkflow_WorkflowNotFound(t *testing.T) {
	start := &mockStartWorkflowRunHandler{}
	getWorkflow := workflowGetter(nil, errors.New("workflow not found"))
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if start.called {
		t.Fatal("start handler must not be called when workflow not found")
	}
}

func TestWorkflowRunHandler_StartWorkflow_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	start := &mockStartWorkflowRunHandler{}
	getWorkflow := workflowGetter(view, nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if start.called {
		t.Fatal("start handler must not be called for wrong project")
	}
}

func TestWorkflowRunHandler_StartWorkflow_GetWorkflowInternalError(t *testing.T) {
	start := &mockStartWorkflowRunHandler{}
	getWorkflow := workflowGetter(nil, errors.New("database unavailable"))
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_StartWorkflow_HandlerError_WorkflowNotFound(t *testing.T) {
	start := &mockStartWorkflowRunHandler{err: domainworkflowrun.ErrWorkflowNotFound}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowRunHandler_StopWorkflow_GetWorkflowInternalError(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{}
	getWorkflow := workflowGetter(nil, errors.New("database unavailable"))
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called when get workflow fails")
	}
}

func TestWorkflowRunHandler_StartWorkflow_AlreadyInProgress(t *testing.T) {
	start := &mockStartWorkflowRunHandler{err: domainworkflowrun.ErrAlreadyInProgress}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}
	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "RUN_IN_PROGRESS" {
		t.Fatalf("code: got %v", body["code"])
	}
}

func TestWorkflowRunHandler_StartWorkflow_QuotaExceeded(t *testing.T) {
	start := &mockStartWorkflowRunHandler{err: cmdquota.ErrWorkflowRunQuotaExceeded}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestWorkflowRunHandler_StartWorkflow_HandlerError_Internal(t *testing.T) {
	start := &mockStartWorkflowRunHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(start, nil, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/start", activeProject(), h.StartWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/start", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_StopWorkflow_Success(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{result: sampleWorkflowRunEntity()}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !cancel.called {
		t.Fatal("expected cancel handler to be called")
	}
	if cancel.cmd.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", cancel.cmd.WorkflowID)
	}
	if cancel.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", cancel.cmd.ProjectID)
	}
	if cancel.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", cancel.cmd.UserID)
	}
}

func TestWorkflowRunHandler_StopWorkflow_Unauthorized(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{}
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called without user")
	}
}

func TestWorkflowRunHandler_StopWorkflow_MissingActiveProject(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{}
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", testutil.WithUserWithoutProject(testutil.TestUserID), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called without active project")
	}
}

func TestWorkflowRunHandler_StopWorkflow_InvalidWorkflowID(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{}
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/bad-id/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called with invalid workflow id")
	}
}

func TestWorkflowRunHandler_StopWorkflow_WorkflowNotFound(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{}
	getWorkflow := workflowGetter(nil, errors.New("workflow not found"))
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called when workflow not found")
	}
}

func TestWorkflowRunHandler_StopWorkflow_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	cancel := &mockCancelWorkflowRunHandler{}
	getWorkflow := workflowGetter(view, nil)
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if cancel.called {
		t.Fatal("cancel handler must not be called for wrong project")
	}
}

func TestWorkflowRunHandler_StopWorkflow_NoRunInProgress(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{err: domainworkflowrun.ErrNoRunInProgress}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}
	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "NO_RUN_IN_PROGRESS" {
		t.Fatalf("code: got %v", body["code"])
	}
}

func TestWorkflowRunHandler_StopWorkflow_CancelWorkflowNotFound(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{err: domainworkflowrun.ErrWorkflowNotFound}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowRunHandler_StopWorkflow_HandlerError_Internal(t *testing.T) {
	cancel := &mockCancelWorkflowRunHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, cancel, nil, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/stop", activeProject(), h.StopWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/stop", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_ListByWorkflow_Success(t *testing.T) {
	runView := sampleWorkflowRunView()
	list := &mockListWorkflowRunsByWorkflowHandler{
		views: []domainworkflowrun.WorkflowRunView{*runView},
		total: 1,
	}
	listStepRunsByIDs := &mockListStepRunsByWorkflowRunIDsHandler{
		views: []domainsteprun.StepRunView{sampleStepRunView()},
	}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, listStepRunsByIDs, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs?page=1&limit=10", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if !listStepRunsByIDs.called {
		t.Fatal("expected list step runs by ids handler to be called")
	}
	if list.query.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", list.query.WorkflowID)
	}
	if list.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", list.query.UserID)
	}
	if list.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", list.query.ProjectID)
	}

	var out struct {
		Members []presenter.WorkflowRunListResponse `json:"members"`
		Total   int                                 `json:"total"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if len(out.Members) != 1 {
		t.Fatalf("members length: got %d want 1", len(out.Members))
	}
	if len(out.Members[0].StepRuns) != 1 {
		t.Fatalf("step runs length: got %d want 1", len(out.Members[0].StepRuns))
	}
}

func TestWorkflowRunHandler_ListByWorkflow_MissingActiveProject(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{}
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", testutil.WithUserWithoutProject(testutil.TestUserID), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called {
		t.Fatal("list handler must not be called without active project")
	}
}

func TestWorkflowRunHandler_ListByWorkflow_InvalidWorkflowID(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{}
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/bad-id/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called {
		t.Fatal("list handler must not be called with invalid workflow id")
	}
}

func TestWorkflowRunHandler_ListByWorkflow_WorkflowNotFound(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{}
	getWorkflow := workflowGetter(nil, errors.New("workflow not found"))
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if list.called {
		t.Fatal("list handler must not be called when workflow not found")
	}
}

func TestWorkflowRunHandler_ListByWorkflow_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	list := &mockListWorkflowRunsByWorkflowHandler{}
	getWorkflow := workflowGetter(view, nil)
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if list.called {
		t.Fatal("list handler must not be called for wrong project")
	}
}

func TestWorkflowRunHandler_ListByWorkflow_InvalidQuery(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs?page=not-a-number", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called {
		t.Fatal("list handler must not be called with invalid query")
	}
}

func TestWorkflowRunHandler_ListByWorkflow_HandlerError_Internal(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_ListByWorkflow_StepRunsError_Internal(t *testing.T) {
	runView := sampleWorkflowRunView()
	list := &mockListWorkflowRunsByWorkflowHandler{
		views: []domainworkflowrun.WorkflowRunView{*runView},
		total: 1,
	}
	listStepRunsByIDs := &mockListStepRunsByWorkflowRunIDsHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, listStepRunsByIDs, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_GetByID_Success(t *testing.T) {
	runView := sampleWorkflowRunView()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	listStepRuns := &mockListStepRunsByWorkflowRunHandler{views: []domainsteprun.StepRunView{sampleStepRunView()}}
	listInsights := &mockListInsightsByStepRunIDsHandler{views: []domaininsight.InsightView{sampleInsightView()}}
	listSteps := &mockListStepsByWorkflowHandler{views: []domainstep.StepView{sampleStepView()}}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, listStepRuns, nil, listInsights, getWorkflow, listSteps, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !getByID.called || !listStepRuns.called || !listInsights.called || !listSteps.called {
		t.Fatal("expected all relation handlers to be called")
	}

	var out presenter.WorkflowRunDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestWorkflowRunID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
	if len(out.StepRuns) != 1 {
		t.Fatalf("step runs length: got %d want 1", len(out.StepRuns))
	}
}

func TestWorkflowRunHandler_GetByID_InsightsNotAllowed(t *testing.T) {
	runView := sampleWorkflowRunView()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	listStepRuns := &mockListStepRunsByWorkflowRunHandler{views: []domainsteprun.StepRunView{sampleStepRunView()}}
	listInsights := &mockListInsightsByStepRunIDsHandler{views: []domaininsight.InsightView{sampleInsightView()}}
	listSteps := &mockListStepsByWorkflowHandler{views: []domainstep.StepView{sampleStepView()}}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	insightsAllowed := &mockInsightsAllowedChecker{allowed: false}
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, listStepRuns, nil, listInsights, getWorkflow, listSteps, insightsAllowed)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !insightsAllowed.called {
		t.Fatal("expected insights allowed check")
	}
	if listInsights.called {
		t.Fatal("insights must not be loaded when plan disallows insights")
	}

	var out presenter.WorkflowRunDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if len(out.StepRuns) != 1 {
		t.Fatalf("step runs length: got %d want 1", len(out.StepRuns))
	}
	if len(out.StepRuns[0].Insights) != 0 {
		t.Fatalf("insights length: got %d want 0", len(out.StepRuns[0].Insights))
	}
}

func TestWorkflowRunHandler_GetByID_MissingActiveProject(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{}
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called without active project")
	}
}

func TestWorkflowRunHandler_GetByID_InvalidWorkflowID(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{}
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/bad-id/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called with invalid workflow id")
	}
}

func TestWorkflowRunHandler_GetByID_InvalidRunID(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/bad-id"
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called with invalid run id")
	}
}

func TestWorkflowRunHandler_GetByID_WorkflowNotFound(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{}
	getWorkflow := workflowGetter(nil, errors.New("workflow not found"))
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called when workflow not found")
	}
}

func TestWorkflowRunHandler_GetByID_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowRunByIDHandler{}
	getWorkflow := workflowGetter(view, nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called for wrong project")
	}
}

func TestWorkflowRunHandler_GetByID_RunNotFound(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{err: errors.New("workflow run not found")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowRunHandler_GetByID_RunWrongWorkflow(t *testing.T) {
	runView := sampleWorkflowRunView()
	runView.WorkflowID = uuid.New()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowRunHandler_GetByID_GetRunInternalError(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_GetByID_StepRunsError_Internal(t *testing.T) {
	runView := sampleWorkflowRunView()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	listStepRuns := &mockListStepRunsByWorkflowRunHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, listStepRuns, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_GetByID_InsightsError_Internal(t *testing.T) {
	runView := sampleWorkflowRunView()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	listStepRuns := &mockListStepRunsByWorkflowRunHandler{views: []domainsteprun.StepRunView{sampleStepRunView()}}
	listInsights := &mockListInsightsByStepRunIDsHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, listStepRuns, nil, listInsights, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_GetByID_StepsError_Internal(t *testing.T) {
	runView := sampleWorkflowRunView()
	getByID := &mockGetWorkflowRunByIDHandler{view: runView}
	listStepRuns := &mockListStepRunsByWorkflowRunHandler{views: []domainsteprun.StepRunView{sampleStepRunView()}}
	listInsights := &mockListInsightsByStepRunIDsHandler{views: []domaininsight.InsightView{sampleInsightView()}}
	listSteps := &mockListStepsByWorkflowHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, listStepRuns, nil, listInsights, getWorkflow, listSteps, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_ListByWorkflow_GetWorkflowInternalError(t *testing.T) {
	list := &mockListWorkflowRunsByWorkflowHandler{}
	getWorkflow := workflowGetter(nil, errors.New("database unavailable"))
	h := newWorkflowRunHandler(nil, nil, nil, nil, list, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/runs", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_GetByID_GetWorkflowInternalError(t *testing.T) {
	getByID := &mockGetWorkflowRunByIDHandler{}
	getWorkflow := workflowGetter(nil, errors.New("database unavailable"))
	h := newWorkflowRunHandler(nil, nil, getByID, nil, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/" + testutil.TestWorkflowRunID.String()
	app.Get("/workflows/:workflowId/runs/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called when get workflow fails")
	}
}

func TestWorkflowRunHandler_Analytics_Success(t *testing.T) {
	lastRun := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	analytics := &mockWorkflowRunAnalyticsHandler{
		stats: &domainworkflowrun.WorkflowRunAnalytics{
			TotalRuns:         10,
			SuccessRate:       0.8,
			SuccessCount:      8,
			FailureRate:       0.2,
			FailureCount:      2,
			AverageDurationMS: 1500,
			LastRunAt:         &lastRun,
		},
	}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	from := "2026-01-01T00:00:00Z"
	to := "2026-01-31T23:59:59Z"
	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics?from=" + from + "&to=" + to
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !analytics.called {
		t.Fatal("expected analytics handler to be called")
	}
	if analytics.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", analytics.query.ProjectID)
	}
	if analytics.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", analytics.query.UserID)
	}
	if analytics.query.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", analytics.query.WorkflowID)
	}

	var out presenter.WorkflowRunAnalyticsResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.TotalRuns != 10 {
		t.Fatalf("total runs: got %d want 10", out.TotalRuns)
	}
}

func TestWorkflowRunHandler_Analytics_MissingActiveProject(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", testutil.WithUserWithoutProject(testutil.TestUserID), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called without active project")
	}
}

func TestWorkflowRunHandler_Analytics_InvalidWorkflowID(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/bad-id/runs/analytics", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called with invalid workflow id")
	}
}

func TestWorkflowRunHandler_Analytics_WorkflowNotFound(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	getWorkflow := workflowGetter(nil, errors.New("workflow not found"))
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called when workflow not found")
	}
}

func TestWorkflowRunHandler_Analytics_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	analytics := &mockWorkflowRunAnalyticsHandler{}
	getWorkflow := workflowGetter(view, nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called for wrong project")
	}
}

func TestWorkflowRunHandler_Analytics_InvalidFromDate(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics?from=not-rfc3339"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called with invalid from date")
	}
}

func TestWorkflowRunHandler_Analytics_InvalidToDate(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics?to=not-rfc3339"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if analytics.called {
		t.Fatal("analytics handler must not be called with invalid to date")
	}
}

func TestWorkflowRunHandler_Analytics_FromAfterTo(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{err: errors.New("from must be before to")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics?from=2026-02-01T00:00:00Z&to=2026-01-01T00:00:00Z"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowRunHandler_Analytics_HandlerError_Internal(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{err: errors.New("database unavailable")}
	getWorkflow := workflowGetter(sampleWorkflowView(), nil)
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowRunHandler_Analytics_GetWorkflowInternalError(t *testing.T) {
	analytics := &mockWorkflowRunAnalyticsHandler{}
	getWorkflow := workflowGetter(nil, errors.New("database unavailable"))
	h := newWorkflowRunHandler(nil, nil, nil, analytics, nil, nil, nil, nil, getWorkflow, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/analytics", activeProject(), h.Analytics)

	path := "/workflows/" + testutil.TestWorkflowID.String() + "/runs/analytics"
	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
