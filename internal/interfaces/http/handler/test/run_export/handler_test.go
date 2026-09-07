package runexporttest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	runexportcmd "go-api/internal/application/command/runexport"
	queryrunexport "go-api/internal/application/query/runexport"
	queryworkflow "go-api/internal/application/query/workflow"
	domainrunexport "go-api/internal/domain/runexport"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockRequestRunExportHandler struct {
	called bool
	cmd    runexportcmd.RequestRunExportCommand
	result *domainrunexport.RunExport
	err    error
}

func (m *mockRequestRunExportHandler) Handle(
	_ context.Context,
	cmd runexportcmd.RequestRunExportCommand,
) (*domainrunexport.RunExport, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockGetRunExportByIDHandler struct {
	called bool
	query  queryrunexport.GetRunExportByIDQuery
	view   *domainrunexport.View
	err    error
}

func (m *mockGetRunExportByIDHandler) Handle(
	_ context.Context,
	q queryrunexport.GetRunExportByIDQuery,
) (*domainrunexport.View, error) {
	m.called = true
	m.query = q
	return m.view, m.err
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

func newRunExportHandler(
	request *mockRequestRunExportHandler,
	getByID *mockGetRunExportByIDHandler,
	getWorkflow *mockGetWorkflowByIDHandler,
) *handler.RunExportHandler {
	if request == nil {
		request = &mockRequestRunExportHandler{}
	}
	if getByID == nil {
		getByID = &mockGetRunExportByIDHandler{}
	}
	if getWorkflow == nil {
		getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewRunExportHandler(request, getByID, getWorkflow)
}

func sampleWorkflowView() *domainworkflow.WorkflowView {
	return &domainworkflow.WorkflowView{
		ID:        testutil.TestWorkflowID,
		Name:      "Order Flow",
		ProjectID: testutil.TestProjectID,
		Status:    domainworkflow.StatusActive,
	}
}

func sampleRunExport() *domainrunexport.RunExport {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	return &domainrunexport.RunExport{
		ID:                testutil.TestRunExportID,
		WorkflowID:        testutil.TestWorkflowID,
		ProjectID:         testutil.TestProjectID,
		RequestedByUserID: testutil.TestUserID,
		DateRangeFrom:     from,
		DateRangeTo:       to,
		Status:            domainrunexport.StatusPending,
		CreatedAt:         from,
		UpdatedAt:         from,
	}
}

func sampleRunExportView() *domainrunexport.View {
	from := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)
	return &domainrunexport.View{
		ID:                testutil.TestRunExportID,
		WorkflowID:        testutil.TestWorkflowID,
		ProjectID:         testutil.TestProjectID,
		RequestedByUserID: testutil.TestUserID,
		DateRangeFrom:     from,
		DateRangeTo:       to,
		Status:            domainrunexport.StatusReady,
		CreatedAt:         from,
		UpdatedAt:         to,
	}
}

func exportPath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/runs/export"
}

func exportJobPath() string {
	return exportPath() + "/" + testutil.TestRunExportID.String()
}

func mustJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	req, err := testutil.JSONRequest(method, path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	return req
}

func TestRunExportHandler_Create_Success(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{result: sampleRunExport()}
	h := newRunExportHandler(request, nil, getWorkflow)

	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusAccepted)
	}
	var out presenter.RunExportAcceptedResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestRunExportID.String() {
		t.Fatalf("id: got %s", out.ID)
	}
	if out.Status != string(domainrunexport.StatusPending) {
		t.Fatalf("status: got %s", out.Status)
	}
	if !request.called {
		t.Fatal("expected request handler to be called")
	}
	if request.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", request.cmd.UserID)
	}
	if request.cmd.WorkflowID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", request.cmd.WorkflowID)
	}
}

func TestRunExportHandler_Create_Unauthorized(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRunExportHandler_Create_MissingActiveProject(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_Create_InvalidWorkflowID(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/not-a-uuid/runs/export", map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_Create_WorkflowNotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	h := newRunExportHandler(nil, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_Create_WorkflowLookupFailed(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newRunExportHandler(nil, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestRunExportHandler_Create_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if request.called {
		t.Fatal("request handler must not be called")
	}
}

func TestRunExportHandler_Create_InvalidBody(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := http.NewRequest(http.MethodPost, exportPath(), bytes.NewBufferString("not-json"))
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
	if request.called {
		t.Fatal("request handler must not be called with invalid body")
	}
}

func TestRunExportHandler_Create_EmptyBody(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{result: sampleRunExport()}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusAccepted)
	}
}

func TestRunExportHandler_Create_InvalidDateRange(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{err: domainrunexport.ErrInvalidDateRange}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_Create_NotAllowed(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{err: cmdquota.ErrDataExportNotAllowed}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestRunExportHandler_Create_HandlerError_NotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{err: errors.New("workflow not found")}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_Create_HandlerError_Internal(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	request := &mockRequestRunExportHandler{err: errors.New("failed to request run export")}
	h := newRunExportHandler(request, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Post("/workflows/:workflowId/runs/export", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, exportPath(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestRunExportHandler_GetByID_Success(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	getByID := &mockGetRunExportByIDHandler{view: sampleRunExportView()}
	h := newRunExportHandler(nil, getByID, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	var out presenter.RunExportResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestRunExportID.String() {
		t.Fatalf("id: got %s", out.ID)
	}
	if out.Status != string(domainrunexport.StatusReady) {
		t.Fatalf("status: got %s", out.Status)
	}
}

func TestRunExportHandler_GetByID_Unauthorized(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestRunExportHandler_GetByID_MissingActiveProject(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_GetByID_InvalidWorkflowID(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/not-a-uuid/runs/export/"+testutil.TestRunExportID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_GetByID_InvalidJobID(t *testing.T) {
	h := newRunExportHandler(nil, nil, nil)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportPath()+"/not-a-uuid", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRunExportHandler_GetByID_WorkflowNotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	h := newRunExportHandler(nil, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_GetByID_WorkflowLookupFailed(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newRunExportHandler(nil, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestRunExportHandler_GetByID_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newRunExportHandler(nil, nil, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_GetByID_NotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	getByID := &mockGetRunExportByIDHandler{err: errors.New("run export not found")}
	h := newRunExportHandler(nil, getByID, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_GetByID_HandlerError_Internal(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	getByID := &mockGetRunExportByIDHandler{err: errors.New("failed to get run export")}
	h := newRunExportHandler(nil, getByID, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestRunExportHandler_GetByID_JobWrongProject(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	view := sampleRunExportView()
	view.ProjectID = otherProjectID
	getByID := &mockGetRunExportByIDHandler{view: view}
	h := newRunExportHandler(nil, getByID, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestRunExportHandler_GetByID_WrongWorkflow(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	view := sampleRunExportView()
	view.WorkflowID = uuid.MustParse("01960000-0000-7000-8000-0000000000ee")
	getByID := &mockGetRunExportByIDHandler{view: view}
	h := newRunExportHandler(nil, getByID, getWorkflow)
	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/runs/export/:jobId", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, exportJobPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}
