package assertiontest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	assertioncmd "go-api/internal/application/command/assertion"
	queryassertion "go-api/internal/application/query/assertion"
	querystep "go-api/internal/application/query/step"
	queryworkflow "go-api/internal/application/query/workflow"
	domainassertion "go-api/internal/domain/assertion"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockCreateAssertionHandler struct {
	called bool
	cmd    assertioncmd.CreateAssertionCommand
	result *domainassertion.Assertion
	err    error
}

func (m *mockCreateAssertionHandler) Handle(
	_ context.Context,
	cmd assertioncmd.CreateAssertionCommand,
) (*domainassertion.Assertion, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateAssertionHandler struct {
	called bool
	cmd    assertioncmd.UpdateAssertionCommand
	result *domainassertion.Assertion
	err    error
}

func (m *mockUpdateAssertionHandler) Handle(
	_ context.Context,
	cmd assertioncmd.UpdateAssertionCommand,
) (*domainassertion.Assertion, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeleteAssertionHandler struct {
	called bool
	cmd    assertioncmd.DeleteAssertionCommand
	err    error
}

func (m *mockDeleteAssertionHandler) Handle(_ context.Context, cmd assertioncmd.DeleteAssertionCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockGetAssertionByIDHandler struct {
	called bool
	query  queryassertion.GetAssertionByIDQuery
	view   *domainassertion.AssertionView
	err    error
}

func (m *mockGetAssertionByIDHandler) Handle(
	_ context.Context,
	q queryassertion.GetAssertionByIDQuery,
) (*domainassertion.AssertionView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockListAssertionsByStepHandler struct {
	called bool
	query  queryassertion.ListAssertionsByStepQuery
	views  []domainassertion.AssertionView
	err    error
}

func (m *mockListAssertionsByStepHandler) Handle(
	_ context.Context,
	q queryassertion.ListAssertionsByStepQuery,
) ([]domainassertion.AssertionView, error) {
	m.called = true
	m.query = q
	return m.views, m.err
}

type mockSearchAssertionPathsHandler struct {
	called bool
	query  queryassertion.SearchAssertionPathsQuery
	paths  []string
	total  int
	err    error
}

func (m *mockSearchAssertionPathsHandler) Handle(
	_ context.Context,
	q queryassertion.SearchAssertionPathsQuery,
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

type assertionMocks struct {
	create      *mockCreateAssertionHandler
	update      *mockUpdateAssertionHandler
	deleteH     *mockDeleteAssertionHandler
	getByID     *mockGetAssertionByIDHandler
	listByStep  *mockListAssertionsByStepHandler
	searchPaths *mockSearchAssertionPathsHandler
	getStep     *mockGetStepByIDHandler
	getWorkflow *mockGetWorkflowByIDHandler
}

func newAssertionHandler(m assertionMocks) *handler.AssertionHandler {
	if m.create == nil {
		m.create = &mockCreateAssertionHandler{}
	}
	if m.update == nil {
		m.update = &mockUpdateAssertionHandler{}
	}
	if m.deleteH == nil {
		m.deleteH = &mockDeleteAssertionHandler{}
	}
	if m.getByID == nil {
		m.getByID = &mockGetAssertionByIDHandler{}
	}
	if m.listByStep == nil {
		m.listByStep = &mockListAssertionsByStepHandler{}
	}
	if m.searchPaths == nil {
		m.searchPaths = &mockSearchAssertionPathsHandler{}
	}
	if m.getStep == nil {
		m.getStep = &mockGetStepByIDHandler{}
	}
	if m.getWorkflow == nil {
		m.getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewAssertionHandler(
		m.create,
		m.update,
		m.deleteH,
		m.getByID,
		m.listByStep,
		m.searchPaths,
		m.getStep,
		m.getWorkflow,
	)
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func assertionsRoute() string     { return "/workflows/:workflowId/steps/:stepId/assertions" }
func assertionItemRoute() string  { return "/workflows/:workflowId/assertions/:id" }
func assertionPathsRoute() string { return "/workflows/:workflowId/steps/:stepId/assertion-paths" }

func assertionsBasePath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/steps/" + testutil.TestStepID.String() + "/assertions"
}

func assertionItemPath(id uuid.UUID) string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/assertions/" + id.String()
}

func assertionPathsPath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/steps/" + testutil.TestStepID.String() + "/assertion-paths"
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

func stepMocksOK() *mockGetStepByIDHandler {
	return &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleStepView()},
		errs:  []error{nil},
	}
}

func sampleAssertionEntity() *domainassertion.Assertion {
	return &domainassertion.Assertion{
		ID:            testutil.TestAssertionID,
		Description:   "Status is OK",
		Source:        domainassertion.SourceStatus,
		Path:          "",
		Operator:      domainassertion.OperatorEquals,
		ExpectedValue: "200",
		StepID:        testutil.TestStepID,
		WorkflowID:    testutil.TestWorkflowID,
		CreatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleAssertionView() *domainassertion.AssertionView {
	e := sampleAssertionEntity()
	return &domainassertion.AssertionView{
		ID:            e.ID,
		Description:   e.Description,
		Source:        e.Source,
		Path:          e.Path,
		Operator:      e.Operator,
		ExpectedValue: e.ExpectedValue,
		StepID:        e.StepID,
		WorkflowID:    e.WorkflowID,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

func validCreateAssertionBody() map[string]any {
	return map[string]any{
		"description":   "Status is OK",
		"source":        "status",
		"operator":      "equals",
		"expectedValue": "200",
	}
}

func validUpdateAssertionBody() map[string]any {
	return map[string]any{
		"description":   "Status is created",
		"source":        "status",
		"operator":      "equals",
		"expectedValue": "201",
	}
}

func validNotNullAssertionBody() map[string]any {
	return map[string]any{
		"source":   "body",
		"path":     "$.id",
		"operator": "not_null",
	}
}
