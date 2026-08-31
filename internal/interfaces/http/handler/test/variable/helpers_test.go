package variabletest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	variablecmd "go-api/internal/application/command/variable"
	querystep "go-api/internal/application/query/step"
	queryvariable "go-api/internal/application/query/variable"
	queryworkflow "go-api/internal/application/query/workflow"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockCreateVariableHandler struct {
	called bool
	cmd    variablecmd.CreateVariableCommand
	result *domainvariable.Variable
	err    error
}

func (m *mockCreateVariableHandler) Handle(
	_ context.Context,
	cmd variablecmd.CreateVariableCommand,
) (*domainvariable.Variable, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateVariableHandler struct {
	called bool
	cmd    variablecmd.UpdateVariableCommand
	result *domainvariable.Variable
	err    error
}

func (m *mockUpdateVariableHandler) Handle(
	_ context.Context,
	cmd variablecmd.UpdateVariableCommand,
) (*domainvariable.Variable, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeleteVariableHandler struct {
	called bool
	cmd    variablecmd.DeleteVariableCommand
	err    error
}

func (m *mockDeleteVariableHandler) Handle(_ context.Context, cmd variablecmd.DeleteVariableCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockGetVariableByIDHandler struct {
	called bool
	query  queryvariable.GetVariableByIDQuery
	view   *domainvariable.VariableView
	err    error
}

func (m *mockGetVariableByIDHandler) Handle(
	_ context.Context,
	q queryvariable.GetVariableByIDQuery,
) (*domainvariable.VariableView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockListVariablesByWorkflowHandler struct {
	called bool
	query  queryvariable.ListVariablesByWorkflowQuery
	views  []domainvariable.VariableView
	err    error
}

func (m *mockListVariablesByWorkflowHandler) Handle(
	_ context.Context,
	q queryvariable.ListVariablesByWorkflowQuery,
) ([]domainvariable.VariableView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockListAvailableVariablesHandler struct {
	called bool
	query  queryvariable.ListAvailableVariablesQuery
	views  []domainvariable.VariableView
	err    error
}

func (m *mockListAvailableVariablesHandler) Handle(
	_ context.Context,
	q queryvariable.ListAvailableVariablesQuery,
) ([]domainvariable.VariableView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockSearchVariablePathsHandler struct {
	called bool
	query  queryvariable.SearchVariablePathsQuery
	paths  []string
	total  int
	err    error
}

func (m *mockSearchVariablePathsHandler) Handle(
	_ context.Context,
	q queryvariable.SearchVariablePathsQuery,
) ([]string, int, error) {
	m.called = true
	m.query = q
	return m.paths, m.total, m.err
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

type variableMocks struct {
	create         *mockCreateVariableHandler
	update         *mockUpdateVariableHandler
	deleteH        *mockDeleteVariableHandler
	getByID        *mockGetVariableByIDHandler
	listByWorkflow *mockListVariablesByWorkflowHandler
	listAvailable  *mockListAvailableVariablesHandler
	searchPaths    *mockSearchVariablePathsHandler
	getStep        *mockGetStepByIDHandler
	getWorkflow    *mockGetWorkflowByIDHandler
}

func newVariableHandler(m variableMocks) *handler.VariableHandler {
	if m.create == nil {
		m.create = &mockCreateVariableHandler{}
	}
	if m.update == nil {
		m.update = &mockUpdateVariableHandler{}
	}
	if m.deleteH == nil {
		m.deleteH = &mockDeleteVariableHandler{}
	}
	if m.getByID == nil {
		m.getByID = &mockGetVariableByIDHandler{}
	}
	if m.listByWorkflow == nil {
		m.listByWorkflow = &mockListVariablesByWorkflowHandler{}
	}
	if m.listAvailable == nil {
		m.listAvailable = &mockListAvailableVariablesHandler{}
	}
	if m.searchPaths == nil {
		m.searchPaths = &mockSearchVariablePathsHandler{}
	}
	if m.getStep == nil {
		m.getStep = &mockGetStepByIDHandler{}
	}
	if m.getWorkflow == nil {
		m.getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewVariableHandler(
		m.create,
		m.update,
		m.deleteH,
		m.getByID,
		m.listByWorkflow,
		m.listAvailable,
		m.searchPaths,
		m.getStep,
		m.getWorkflow,
	)
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func variablesRoute() string             { return "/workflows/:workflowId/variables" }
func variableItemRoute() string          { return "/workflows/:workflowId/variables/:id" }
func availableVariablesRoute() string    { return "/workflows/:workflowId/steps/:stepId/variables" }
func variablePathsRoute() string         { return "/workflows/:workflowId/steps/:stepId/variable-paths" }

func variablesBasePath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/variables"
}

func variableItemPath(id uuid.UUID) string {
	return variablesBasePath() + "/" + id.String()
}

func availableVariablesPath(stepID uuid.UUID) string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/steps/" + stepID.String() + "/variables"
}

func variablePathsPath(stepID uuid.UUID) string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/steps/" + stepID.String() + "/variable-paths"
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

func workflowMocksOK() *mockGetWorkflowByIDHandler {
	return &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
}

func sampleStepView() *domainstep.StepView {
	endpointID := testutil.TestEndpointID
	return &domainstep.StepView{
		ID:         testutil.TestStepID,
		WorkflowID: testutil.TestWorkflowID,
		EndpointID: &endpointID,
		ProjectID:  testutil.TestProjectID,
	}
}

func sampleVariableEntity() *domainvariable.Variable {
	return &domainvariable.Variable{
		ID:          testutil.TestVariableID,
		Name:        "User ID",
		Key:         "userId",
		Description: "Extracted user id",
		Kind:        domainvariable.KindExtracted,
		Path:        "$.id",
		WorkflowID:  testutil.TestWorkflowID,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleVariableView() *domainvariable.VariableView {
	e := sampleVariableEntity()
	return &domainvariable.VariableView{
		ID:          e.ID,
		Name:        e.Name,
		Key:         e.Key,
		Description: e.Description,
		Kind:        e.Kind,
		Path:        e.Path,
		Value:       e.Value,
		StepID:      e.StepID,
		WorkflowID:  e.WorkflowID,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

func validCreateVariableBody() map[string]any {
	stepID := testutil.TestStepID.String()
	return map[string]any{
		"stepId":      stepID,
		"kind":        "extracted",
		"name":        "User ID",
		"key":         "userId",
		"description": "Extracted user id",
		"path":        "$.id",
	}
}

func validUpdateVariableBody() map[string]any {
	return map[string]any{
		"name":        "User ID",
		"key":         "userId",
		"description": "Updated",
		"path":        "$.data.id",
	}
}

func validStaticVariableBody() map[string]any {
	return map[string]any{
		"kind":  "static",
		"name":  "API Key",
		"key":   "apiKey",
		"value": "secret",
	}
}
