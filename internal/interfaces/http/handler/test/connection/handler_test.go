package connectiontest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	conncmd "go-api/internal/application/command/connection"
	queryconn "go-api/internal/application/query/connection"
	queryworkflow "go-api/internal/application/query/workflow"
	domainconnection "go-api/internal/domain/connection"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var targetStepID = uuid.MustParse("01960000-0000-7000-8000-00000000000c")

type mockCreateConnectionHandler struct {
	called bool
	cmd    conncmd.CreateConnectionCommand
	result *domainconnection.Connection
	err    error
}

func (m *mockCreateConnectionHandler) Handle(
	_ context.Context,
	cmd conncmd.CreateConnectionCommand,
) (*domainconnection.Connection, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeleteConnectionHandler struct {
	called bool
	cmd    conncmd.DeleteConnectionCommand
	err    error
}

func (m *mockDeleteConnectionHandler) Handle(_ context.Context, cmd conncmd.DeleteConnectionCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockListConnectionsByWorkflowHandler struct {
	called bool
	query  queryconn.ListConnectionsByWorkflowQuery
	views  []domainconnection.ConnectionView
	err    error
}

func (m *mockListConnectionsByWorkflowHandler) Handle(
	_ context.Context,
	q queryconn.ListConnectionsByWorkflowQuery,
) ([]domainconnection.ConnectionView, error) {
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

func newConnectionHandler(
	create *mockCreateConnectionHandler,
	deleteH *mockDeleteConnectionHandler,
	list *mockListConnectionsByWorkflowHandler,
	getWorkflow *mockGetWorkflowByIDHandler,
) *handler.ConnectionHandler {
	if create == nil {
		create = &mockCreateConnectionHandler{}
	}
	if deleteH == nil {
		deleteH = &mockDeleteConnectionHandler{}
	}
	if list == nil {
		list = &mockListConnectionsByWorkflowHandler{}
	}
	if getWorkflow == nil {
		getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewConnectionHandler(create, deleteH, list, getWorkflow)
}

func sampleConnectionEntity(branch *domainconnection.ConditionBranch) *domainconnection.Connection {
	return &domainconnection.Connection{
		ID:           testutil.TestConnectionID,
		WorkflowID:   testutil.TestWorkflowID,
		ProjectID:    testutil.TestProjectID,
		SourceStepID: testutil.TestStepID,
		TargetStepID: targetStepID,
		Branch:       branch,
	}
}

func sampleConnectionView(branch *domainconnection.ConditionBranch) domainconnection.ConnectionView {
	c := sampleConnectionEntity(branch)
	return domainconnection.ConnectionView{
		ID:           c.ID,
		SourceStepID: c.SourceStepID,
		TargetStepID: c.TargetStepID,
		Branch:       c.Branch,
	}
}

func validCreateConnectionBody() map[string]any {
	return map[string]any{
		"sourceStepId": testutil.TestStepID.String(),
		"targetStepId": targetStepID.String(),
	}
}

func connectionBasePath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/connections"
}

func connectionsRoute() string {
	return "/workflows/:workflowId/connections"
}

func connectionItemRoute() string {
	return "/workflows/:workflowId/connections/:id"
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func TestConnectionHandler_Create_Success(t *testing.T) {
	create := &mockCreateConnectionHandler{result: sampleConnectionEntity(nil)}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
	if create.cmd.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", create.cmd.WorkflowID)
	}
	if create.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", create.cmd.ProjectID)
	}
	if create.cmd.Branch != nil {
		t.Fatal("expected no branch for plain connection")
	}

	var out presenter.ConnectionDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestConnectionID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestConnectionHandler_Create_BranchTrue(t *testing.T) {
	branch := domainconnection.ConditionBranchTrue
	create := &mockCreateConnectionHandler{result: sampleConnectionEntity(&branch)}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["branch"] = "true"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if create.cmd.Branch == nil {
		t.Fatal("expected branch to be set")
	}
	if *create.cmd.Branch != domainconnection.ConditionBranchTrue {
		t.Fatalf("branch: got %s want true", *create.cmd.Branch)
	}

	var out presenter.ConnectionDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.Branch == nil || *out.Branch != "true" {
		t.Fatalf("response branch: got %v", out.Branch)
	}
}

func TestConnectionHandler_Create_BranchFalse(t *testing.T) {
	branch := domainconnection.ConditionBranchFalse
	create := &mockCreateConnectionHandler{result: sampleConnectionEntity(&branch)}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["branch"] = "false"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if create.cmd.Branch == nil {
		t.Fatal("expected branch to be set")
	}
	if *create.cmd.Branch != domainconnection.ConditionBranchFalse {
		t.Fatalf("branch: got %s want false", *create.cmd.Branch)
	}
}

func TestConnectionHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if create.called {
		t.Fatal("create handler must not be called without user")
	}
}

func TestConnectionHandler_Create_MissingActiveProject(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called without active project")
	}
}

func TestConnectionHandler_Create_InvalidWorkflowID(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/connections", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/bad-id/connections", validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid workflow id")
	}
}

func TestConnectionHandler_Create_InvalidRequestBody(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	req, err := http.NewRequest(http.MethodPost, connectionBasePath(), bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid body")
	}
}

func TestConnectionHandler_Create_InvalidData(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called on validation error")
	}

	body := testutil.DecodeJSONMap(t, resp)
	errorsMap, ok := body["errors"].(map[string]any)
	if !ok {
		t.Fatalf("expected errors map, got %#v", body)
	}
	if _, ok := errorsMap["sourceStepId"]; !ok {
		t.Fatalf("expected sourceStepId validation error, got %#v", errorsMap)
	}
}

func TestConnectionHandler_Create_InvalidSourceStepID(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["sourceStepId"] = "not-a-uuid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid source step id")
	}
}

func TestConnectionHandler_Create_InvalidTargetStepID(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["targetStepId"] = "not-a-uuid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid target step id")
	}
}

func TestConnectionHandler_Create_InvalidBranch(t *testing.T) {
	create := &mockCreateConnectionHandler{}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["branch"] = "maybe"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid branch")
	}
}

func TestConnectionHandler_Create_HandlerError_ConditionRequiresBranch(t *testing.T) {
	create := &mockCreateConnectionHandler{err: domainconnection.ErrConditionRequiresBranch}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Create_HandlerError_NonConditionBranchForbidden(t *testing.T) {
	create := &mockCreateConnectionHandler{err: domainconnection.ErrNonConditionBranchForbidden}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["branch"] = "true"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Create_HandlerError_StepNotFound(t *testing.T) {
	create := &mockCreateConnectionHandler{err: errors.New("source step not found")}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConnectionHandler_Create_HandlerError_SameStep(t *testing.T) {
	create := &mockCreateConnectionHandler{err: errors.New("source and target steps must be different")}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Create_HandlerError_Internal(t *testing.T) {
	create := &mockCreateConnectionHandler{err: errors.New("database unavailable")}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestConnectionHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
	if deleteH.cmd.ID != testutil.TestConnectionID {
		t.Fatalf("connection id: got %s", deleteH.cmd.ID)
	}
	if deleteH.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", deleteH.cmd.ProjectID)
	}
}

func TestConnectionHandler_Delete_Unauthorized(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called without user")
	}
}

func TestConnectionHandler_Delete_InvalidConnectionID(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/bad-id", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called with invalid connection id")
	}
}

func TestConnectionHandler_Delete_NotFound(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{err: errors.New("connection not found")}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConnectionHandler_Delete_HandlerError_Internal(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{err: errors.New("database unavailable")}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestConnectionHandler_ListByWorkflow_Success(t *testing.T) {
	view := sampleConnectionView(nil)
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{{
			ID:        testutil.TestWorkflowID,
			ProjectID: testutil.TestProjectID,
		}},
		errs: []error{nil},
	}
	list := &mockListConnectionsByWorkflowHandler{views: []domainconnection.ConnectionView{view}}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if list.query.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", list.query.WorkflowID)
	}

	var out []presenter.ConnectionDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if len(out) != 1 {
		t.Fatalf("connections length: got %d want 1", len(out))
	}
}

func TestConnectionHandler_ListByWorkflow_MissingActiveProject(t *testing.T) {
	list := &mockListConnectionsByWorkflowHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called || getWorkflow.calls > 0 {
		t.Fatal("handlers must not be called without active project")
	}
}

func TestConnectionHandler_ListByWorkflow_InvalidWorkflowID(t *testing.T) {
	list := &mockListConnectionsByWorkflowHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/connections", activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/bad-id/connections", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called || getWorkflow.calls > 0 {
		t.Fatal("handlers must not be called with invalid workflow id")
	}
}

func TestConnectionHandler_ListByWorkflow_WorkflowNotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	list := &mockListConnectionsByWorkflowHandler{}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
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

func TestConnectionHandler_ListByWorkflow_WrongProject(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{{
			ID:        testutil.TestWorkflowID,
			ProjectID: uuid.New(),
		}},
		errs: []error{nil},
	}
	list := &mockListConnectionsByWorkflowHandler{}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
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

func TestConnectionHandler_ListByWorkflow_HandlerError_Internal(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{{
			ID:        testutil.TestWorkflowID,
			ProjectID: testutil.TestProjectID,
		}},
		errs: []error{nil},
	}
	list := &mockListConnectionsByWorkflowHandler{err: errors.New("database unavailable")}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestConnectionHandler_Create_HandlerError_TargetStepNotFound(t *testing.T) {
	create := &mockCreateConnectionHandler{err: errors.New("target step not found")}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestConnectionHandler_Create_HandlerError_InvalidBranch(t *testing.T) {
	create := &mockCreateConnectionHandler{err: domainconnection.ErrInvalidBranch}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	body := validCreateConnectionBody()
	body["branch"] = "true"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Create_HandlerError_ConditionOutgoingCount(t *testing.T) {
	create := &mockCreateConnectionHandler{err: domainconnection.ErrConditionOutgoingCount}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Create_HandlerError_ConditionalTargetMultipleParents(t *testing.T) {
	create := &mockCreateConnectionHandler{err: domainconnection.ErrConditionalTargetMultipleParents}
	h := newConnectionHandler(create, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post(connectionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, connectionBasePath(), validCreateConnectionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestConnectionHandler_Delete_MissingActiveProject(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete(connectionItemRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, connectionBasePath()+"/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called without active project")
	}
}

func TestConnectionHandler_Delete_InvalidWorkflowID(t *testing.T) {
	deleteH := &mockDeleteConnectionHandler{}
	h := newConnectionHandler(nil, deleteH, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:workflowId/connections/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/bad-id/connections/"+testutil.TestConnectionID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called with invalid workflow id")
	}
}

func TestConnectionHandler_ListByWorkflow_GetWorkflow_InternalError(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	list := &mockListConnectionsByWorkflowHandler{}
	h := newConnectionHandler(nil, nil, list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get(connectionsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, connectionBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if list.called {
		t.Fatal("list handler must not be called when get workflow fails")
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
