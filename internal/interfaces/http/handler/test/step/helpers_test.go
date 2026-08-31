package steptest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	stepcmd "go-api/internal/application/command/step"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryworkflow "go-api/internal/application/query/workflow"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockCreateStepHandler struct {
	called bool
	cmd    stepcmd.CreateStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockCreateStepHandler) Handle(_ context.Context, cmd stepcmd.CreateStepCommand) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockCreateDelayStepHandler struct {
	called bool
	cmd    stepcmd.CreateDelayStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockCreateDelayStepHandler) Handle(_ context.Context, cmd stepcmd.CreateDelayStepCommand) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockCreateConditionStepHandler struct {
	called bool
	cmd    stepcmd.CreateConditionStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockCreateConditionStepHandler) Handle(
	_ context.Context,
	cmd stepcmd.CreateConditionStepCommand,
) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateStepHandler struct {
	called bool
	cmd    stepcmd.UpdateStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockUpdateStepHandler) Handle(_ context.Context, cmd stepcmd.UpdateStepCommand) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateDelayStepHandler struct {
	called bool
	cmd    stepcmd.UpdateDelayStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockUpdateDelayStepHandler) Handle(
	_ context.Context,
	cmd stepcmd.UpdateDelayStepCommand,
) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateConditionStepHandler struct {
	called bool
	cmd    stepcmd.UpdateConditionStepCommand
	result *domainstep.Step
	err    error
}

func (m *mockUpdateConditionStepHandler) Handle(
	_ context.Context,
	cmd stepcmd.UpdateConditionStepCommand,
) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateStepPositionHandler struct {
	called bool
	cmd    stepcmd.UpdateStepPositionCommand
	result *domainstep.Step
	err    error
}

func (m *mockUpdateStepPositionHandler) Handle(
	_ context.Context,
	cmd stepcmd.UpdateStepPositionCommand,
) (*domainstep.Step, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeleteStepHandler struct {
	called bool
	cmd    stepcmd.DeleteStepCommand
	err    error
}

func (m *mockDeleteStepHandler) Handle(_ context.Context, cmd stepcmd.DeleteStepCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockGetStepByIDHandler struct {
	calls int
	views []*domainstep.StepView
	errs  []error
}

func (m *mockGetStepByIDHandler) Handle(
	_ context.Context,
	_ querystep.GetStepByIDQuery,
) (*domainstep.StepView, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get step call")
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

type mockLatestStepRunStatusHandler struct {
	called bool
	query  querysteprun.GetLatestStepRunStatusesByStepIDsQuery
	result map[uuid.UUID]domainsteprun.Status
	err    error
}

func (m *mockLatestStepRunStatusHandler) Handle(
	_ context.Context,
	q querysteprun.GetLatestStepRunStatusesByStepIDsQuery,
) (map[uuid.UUID]domainsteprun.Status, error) {
	m.called = true
	m.query = q
	if m.result == nil {
		m.result = map[uuid.UUID]domainsteprun.Status{}
	}
	return m.result, m.err
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

type stepMocks struct {
	create              *mockCreateStepHandler
	createDelay         *mockCreateDelayStepHandler
	createCondition     *mockCreateConditionStepHandler
	update              *mockUpdateStepHandler
	updateDelay         *mockUpdateDelayStepHandler
	updateCondition     *mockUpdateConditionStepHandler
	updatePosition      *mockUpdateStepPositionHandler
	deleteH             *mockDeleteStepHandler
	getByID             *mockGetStepByIDHandler
	listByWorkflow      *mockListStepsByWorkflowHandler
	latestRunStatus     *mockLatestStepRunStatusHandler
	getWorkflow         *mockGetWorkflowByIDHandler
}

func newStepHandler(m stepMocks) *handler.StepHandler {
	if m.create == nil {
		m.create = &mockCreateStepHandler{}
	}
	if m.createDelay == nil {
		m.createDelay = &mockCreateDelayStepHandler{}
	}
	if m.createCondition == nil {
		m.createCondition = &mockCreateConditionStepHandler{}
	}
	if m.update == nil {
		m.update = &mockUpdateStepHandler{}
	}
	if m.updateDelay == nil {
		m.updateDelay = &mockUpdateDelayStepHandler{}
	}
	if m.updateCondition == nil {
		m.updateCondition = &mockUpdateConditionStepHandler{}
	}
	if m.updatePosition == nil {
		m.updatePosition = &mockUpdateStepPositionHandler{}
	}
	if m.deleteH == nil {
		m.deleteH = &mockDeleteStepHandler{}
	}
	if m.getByID == nil {
		m.getByID = &mockGetStepByIDHandler{}
	}
	if m.listByWorkflow == nil {
		m.listByWorkflow = &mockListStepsByWorkflowHandler{}
	}
	if m.latestRunStatus == nil {
		m.latestRunStatus = &mockLatestStepRunStatusHandler{}
	}
	if m.getWorkflow == nil {
		m.getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewStepHandler(
		m.create,
		m.createDelay,
		m.createCondition,
		m.update,
		m.updateDelay,
		m.updateCondition,
		m.updatePosition,
		m.deleteH,
		m.getByID,
		m.listByWorkflow,
		m.latestRunStatus,
		m.getWorkflow,
	)
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func stepsRoute() string       { return "/workflows/:workflowId/steps" }
func stepItemRoute() string    { return "/workflows/:workflowId/steps/:id" }
func stepPositionRoute() string { return "/workflows/:workflowId/steps/:id/position" }

func stepsBasePath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/steps"
}

func stepItemPath(id uuid.UUID) string {
	return stepsBasePath() + "/" + id.String()
}

func stepPositionPath(id uuid.UUID) string {
	return stepItemPath(id) + "/position"
}

func mustJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	req, err := testutil.JSONRequest(method, path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	return req
}

func sampleWorkflowView() *domainworkflow.WorkflowView {
	return &domainworkflow.WorkflowView{
		ID:        testutil.TestWorkflowID,
		ProjectID: testutil.TestProjectID,
	}
}

func sampleHTTPStepEntity() *domainstep.Step {
	endpointID := testutil.TestEndpointID
	return &domainstep.Step{
		ID:             testutil.TestStepID,
		WorkflowID:     testutil.TestWorkflowID,
		EndpointID:     &endpointID,
		ProjectID:      testutil.TestProjectID,
		Type:           domainstep.TypeHTTP,
		Name:           "Fetch Users",
		URL:            "https://api.example.com/users",
		Method:         "GET",
		Headers:        map[string]string{},
		Query:          httpquery.Empty(),
		Body:           map[string]any{},
		Timeout:        30000,
		RetryOnFailure: false,
		RetryCount:     0,
		RetryDelay:     10000,
		Status:         domainstep.StatusActive,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleDelayStepEntity() *domainstep.Step {
	return &domainstep.Step{
		ID:                   testutil.TestStepID,
		WorkflowID:           testutil.TestWorkflowID,
		ProjectID:            testutil.TestProjectID,
		Type:                 domainstep.TypeDelay,
		Name:                 "Wait",
		DelayDurationSeconds: 30,
		Headers:              map[string]string{},
		Query:                httpquery.Empty(),
		Body:                 map[string]any{},
		Status:               domainstep.StatusActive,
		CreatedAt:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleConditionStepEntity() *domainstep.Step {
	expression := "status == 200"
	return &domainstep.Step{
		ID:             testutil.TestStepID,
		WorkflowID:     testutil.TestWorkflowID,
		ProjectID:      testutil.TestProjectID,
		Type:           domainstep.TypeCondition,
		Name:           "Check Status",
		Expression:     &expression,
		Headers:        map[string]string{},
		Query:          httpquery.Empty(),
		Body:           map[string]any{},
		Status:         domainstep.StatusActive,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleHTTPStepView() *domainstep.StepView {
	e := sampleHTTPStepEntity()
	return &domainstep.StepView{
		ID:             e.ID,
		WorkflowID:     e.WorkflowID,
		EndpointID:     e.EndpointID,
		ProjectID:      e.ProjectID,
		Type:           e.Type,
		Name:           e.Name,
		URL:            e.URL,
		Method:         e.Method,
		Headers:        e.Headers,
		Query:          e.Query,
		Body:           e.Body,
		Timeout:        e.Timeout,
		RetryOnFailure: e.RetryOnFailure,
		RetryCount:     e.RetryCount,
		RetryDelay:     e.RetryDelay,
		Status:         e.Status,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func sampleDelayStepView() *domainstep.StepView {
	e := sampleDelayStepEntity()
	return &domainstep.StepView{
		ID:                   e.ID,
		WorkflowID:           e.WorkflowID,
		ProjectID:            e.ProjectID,
		Type:                 e.Type,
		Name:                 e.Name,
		DelayDurationSeconds: e.DelayDurationSeconds,
		Headers:              e.Headers,
		Query:                e.Query,
		Body:                 e.Body,
		Status:               e.Status,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}

func sampleConditionStepView() *domainstep.StepView {
	e := sampleConditionStepEntity()
	return &domainstep.StepView{
		ID:             e.ID,
		WorkflowID:     e.WorkflowID,
		ProjectID:      e.ProjectID,
		Type:           e.Type,
		Name:           e.Name,
		Expression:     e.Expression,
		Headers:        e.Headers,
		Query:          e.Query,
		Body:           e.Body,
		Status:         e.Status,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func validCreateHTTPStepBody() map[string]any {
	return map[string]any{
		"endpointId": testutil.TestEndpointID.String(),
		"position":   map[string]any{"x": 10.0, "y": 20.0},
	}
}

func validCreateDelayStepBody() map[string]any {
	return map[string]any{
		"type":                 "delay",
		"name":                 "Wait",
		"delayDurationSeconds": 30,
		"position":             map[string]any{"x": 1.0, "y": 2.0},
	}
}

func validCreateConditionStepBody() map[string]any {
	return map[string]any{
		"type":       "condition",
		"name":       "Check Status",
		"expression": "status == 200",
		"position":   map[string]any{"x": 3.0, "y": 4.0},
	}
}

func validUpdateHTTPStepBody() map[string]any {
	return map[string]any{
		"name":           "Fetch Users",
		"url":            "https://api.example.com/users",
		"method":         "GET",
		"timeout":        30000,
		"retryOnFailure": false,
		"retryCount":     0,
		"retryDelay":     10000,
	}
}

func validUpdateDelayStepBody() map[string]any {
	return map[string]any{
		"name":                 "Wait Longer",
		"description":          "Pause before next step",
		"delayDurationSeconds": 60,
	}
}

func validUpdateConditionStepBody() map[string]any {
	return map[string]any{
		"name":        "Check Status",
		"description": "Validate response",
		"expression":  "status == 201",
	}
}

func validUpdatePositionBody() map[string]any {
	return map[string]any{"position": map[string]any{"x": 50.0, "y": 60.0}}
}

func quotaErr() error { return cmdquota.ErrStepQuotaExceeded }
