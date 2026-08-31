package workflowtest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	workflowcmd "go-api/internal/application/command/workflow"
	queryproject "go-api/internal/application/query/project"
	queryworkflow "go-api/internal/application/query/workflow"
	domainproject "go-api/internal/domain/project"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockCreateWorkflowHandler struct {
	called bool
	cmd    workflowcmd.CreateWorkflowCommand
	result *domainworkflow.Workflow
	err    error
}

func (m *mockCreateWorkflowHandler) Handle(
	_ context.Context,
	cmd workflowcmd.CreateWorkflowCommand,
) (*domainworkflow.Workflow, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateWorkflowHandler struct {
	called bool
	cmd    workflowcmd.UpdateWorkflowCommand
	err    error
}

func (m *mockUpdateWorkflowHandler) Handle(_ context.Context, cmd workflowcmd.UpdateWorkflowCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockActivateWorkflowHandler struct {
	called bool
	cmd    workflowcmd.ActivateWorkflowCommand
	result *domainworkflow.Workflow
	err    error
}

func (m *mockActivateWorkflowHandler) Handle(
	_ context.Context,
	cmd workflowcmd.ActivateWorkflowCommand,
) (*domainworkflow.Workflow, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeactivateWorkflowHandler struct {
	called bool
	cmd    workflowcmd.DeactivateWorkflowCommand
	result *domainworkflow.Workflow
	err    error
}

func (m *mockDeactivateWorkflowHandler) Handle(
	_ context.Context,
	cmd workflowcmd.DeactivateWorkflowCommand,
) (*domainworkflow.Workflow, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockDeleteWorkflowHandler struct {
	called bool
	cmd    workflowcmd.DeleteWorkflowCommand
	err    error
}

func (m *mockDeleteWorkflowHandler) Handle(_ context.Context, cmd workflowcmd.DeleteWorkflowCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
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
		return nil, errors.New("unexpected get by id call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	return m.views[idx], nil
}

type mockListWorkflowsByProjectHandler struct {
	called bool
	query  queryworkflow.ListWorkflowsByProjectQuery
	views  []domainworkflow.WorkflowView
	total  int64
	err    error
}

func (m *mockListWorkflowsByProjectHandler) Handle(
	_ context.Context,
	q queryworkflow.ListWorkflowsByProjectQuery,
) ([]domainworkflow.WorkflowView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockGetProjectByIDHandler struct {
	calls int
	views []*domainproject.ProjectView
	errs  []error
}

func (m *mockGetProjectByIDHandler) Handle(
	_ context.Context,
	_ queryproject.GetProjectByIDQuery,
) (*domainproject.ProjectView, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get project call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	return m.views[idx], nil
}

func newWorkflowHandler(
	create *mockCreateWorkflowHandler,
	update *mockUpdateWorkflowHandler,
	activate *mockActivateWorkflowHandler,
	deactivate *mockDeactivateWorkflowHandler,
	deleteH *mockDeleteWorkflowHandler,
	getByID *mockGetWorkflowByIDHandler,
	list *mockListWorkflowsByProjectHandler,
	getProject *mockGetProjectByIDHandler,
) *handler.WorkflowHandler {
	if create == nil {
		create = &mockCreateWorkflowHandler{}
	}
	if update == nil {
		update = &mockUpdateWorkflowHandler{}
	}
	if activate == nil {
		activate = &mockActivateWorkflowHandler{}
	}
	if deactivate == nil {
		deactivate = &mockDeactivateWorkflowHandler{}
	}
	if deleteH == nil {
		deleteH = &mockDeleteWorkflowHandler{}
	}
	if getByID == nil {
		getByID = &mockGetWorkflowByIDHandler{}
	}
	if list == nil {
		list = &mockListWorkflowsByProjectHandler{}
	}
	if getProject == nil {
		getProject = &mockGetProjectByIDHandler{}
	}
	return handler.NewWorkflowHandler(create, update, activate, deactivate, deleteH, getByID, list, getProject)
}

func sampleWorkflowEntity() *domainworkflow.Workflow {
	return &domainworkflow.Workflow{
		ID:                   testutil.TestWorkflowID,
		Name:                 "Order Flow",
		Description:          "Processes orders",
		Status:               domainworkflow.StatusInactive,
		ProjectID:            testutil.TestProjectID,
		ScheduleType:         domainworkflow.ScheduleTypeNone,
		ScheduleTimezone:     "UTC",
		Concurrency:          1,
		NotificationsEnabled: true,
		NotifyOnSuccess:      true,
		NotifyOnFailure:      true,
		NotifyOnCancel:       true,
		CreatedAt:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:            time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleWorkflowView() *domainworkflow.WorkflowView {
	e := sampleWorkflowEntity()
	return &domainworkflow.WorkflowView{
		ID:                   e.ID,
		Name:                 e.Name,
		Description:          e.Description,
		Status:               e.Status,
		ProjectID:            e.ProjectID,
		ScheduleType:         e.ScheduleType,
		ScheduleTimezone:     e.ScheduleTimezone,
		Concurrency:          e.Concurrency,
		NotificationsEnabled: e.NotificationsEnabled,
		NotifyOnSuccess:      e.NotifyOnSuccess,
		NotifyOnFailure:      e.NotifyOnFailure,
		NotifyOnCancel:       e.NotifyOnCancel,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}

func validCreateWorkflowBody() map[string]any {
	return map[string]any{"name": "Order Flow", "description": "Processes orders"}
}

func validUpdateWorkflowBody() map[string]any {
	return map[string]any{
		"name":         "Order Flow",
		"description":  "Updated",
		"status":       "inactive",
		"scheduleType": "none",
	}
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func TestWorkflowHandler_Create_Success(t *testing.T) {
	create := &mockCreateWorkflowHandler{result: sampleWorkflowEntity()}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
	if create.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", create.cmd.UserID)
	}
	if create.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", create.cmd.ProjectID)
	}

	var out presenter.WorkflowDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestWorkflowID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestWorkflowHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateWorkflowHandler{}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
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

func TestWorkflowHandler_Create_MissingActiveProject(t *testing.T) {
	create := &mockCreateWorkflowHandler{}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
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

func TestWorkflowHandler_Create_InvalidRequestBody(t *testing.T) {
	create := &mockCreateWorkflowHandler{}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	req, err := http.NewRequest(http.MethodPost, "/workflows", bytes.NewBufferString("not-json"))
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

func TestWorkflowHandler_Create_InvalidData(t *testing.T) {
	create := &mockCreateWorkflowHandler{}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called on validation error")
	}
}

func TestWorkflowHandler_Create_HandlerError(t *testing.T) {
	create := &mockCreateWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Create_QuotaExceeded(t *testing.T) {
	create := &mockCreateWorkflowHandler{err: cmdquota.ErrWorkflowQuotaExceeded}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestWorkflowHandler_Create_ScheduleError(t *testing.T) {
	create := &mockCreateWorkflowHandler{err: domainworkflow.ErrInvalidSchedule}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", validCreateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_GetByID_Success(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if getByID.calls != 1 {
		t.Fatalf("get by id calls: got %d want 1", getByID.calls)
	}
}

func TestWorkflowHandler_GetByID_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	getProject := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{{
			ID:        otherProjectID,
			Name:      "Other Project",
			MemberIDs: []uuid.UUID{testutil.TestUserID},
		}},
		errs: []error{nil},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, getProject)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["code"] != "WRONG_ORGANIZATION" {
		t.Fatalf("code: got %v want WRONG_ORGANIZATION", body["code"])
	}
	if body["projectId"] != otherProjectID.String() {
		t.Fatalf("projectId: got %v", body["projectId"])
	}
	if body["projectName"] != "Other Project" {
		t.Fatalf("projectName: got %v", body["projectName"])
	}
}

func TestWorkflowHandler_GetByID_Unauthorized(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if getByID.calls != 0 {
		t.Fatal("get by id handler must not be called without user")
	}
}

func TestWorkflowHandler_GetByID_MissingActiveProject(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getByID.calls != 0 {
		t.Fatal("get by id handler must not be called without active project")
	}
}

func TestWorkflowHandler_GetByID_InvalidID(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/not-a-uuid", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getByID.calls != 0 {
		t.Fatal("get by id handler must not be called with invalid id")
	}
}

func TestWorkflowHandler_GetByID_NotFound(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_GetByID_HandlerError_Internal(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_GetByID_WrongProjectNotMember(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	getProject := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{{
			ID:        otherProjectID,
			Name:      "Other Project",
			MemberIDs: []uuid.UUID{uuid.New()},
		}},
		errs: []error{nil},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, getProject)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_GetByID_WrongProjectLookupFailed(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	getProject := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{nil},
		errs:  []error{errors.New("project not found")},
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, getByID, nil, getProject)

	app := testutil.NewTestApp()
	app.Get("/workflows/:id", activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_ListByProject_Success(t *testing.T) {
	view := sampleWorkflowView()
	list := &mockListWorkflowsByProjectHandler{
		views: []domainworkflow.WorkflowView{*view},
		total: 1,
	}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows", activeProject(), h.ListByProject)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows?page=1&limit=10", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if list.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", list.query.ProjectID)
	}
}

func TestWorkflowHandler_ListByProject_MissingActiveProject(t *testing.T) {
	list := &mockListWorkflowsByProjectHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows", testutil.WithUserWithoutProject(testutil.TestUserID), h.ListByProject)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows", nil))
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

func TestWorkflowHandler_ListByProject_InvalidQuery(t *testing.T) {
	list := &mockListWorkflowsByProjectHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows", activeProject(), h.ListByProject)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows?page=not-a-number", nil))
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

func TestWorkflowHandler_ListByProject_HandlerError(t *testing.T) {
	list := &mockListWorkflowsByProjectHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(nil, nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows", activeProject(), h.ListByProject)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Update_Success(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view, view},
		errs:  []error{nil, nil},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !update.called {
		t.Fatal("expected update handler to be called")
	}
	if update.cmd.ID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", update.cmd.ID)
	}
	if getByID.calls != 2 {
		t.Fatalf("get by id calls: got %d want 2", getByID.calls)
	}
}

func TestWorkflowHandler_Update_Unauthorized(t *testing.T) {
	update := &mockUpdateWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if update.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called without user")
	}
}

func TestWorkflowHandler_Update_MissingActiveProject(t *testing.T) {
	update := &mockUpdateWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called without active project")
	}
}

func TestWorkflowHandler_Update_InvalidID(t *testing.T) {
	update := &mockUpdateWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/bad-id", validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called with invalid id")
	}
}

func TestWorkflowHandler_Update_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if update.called {
		t.Fatal("update handler must not be called for wrong project")
	}
}

func TestWorkflowHandler_Update_InvalidRequestBody(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	req, err := http.NewRequest(http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), bytes.NewBufferString("not-json"))
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
	if update.called {
		t.Fatal("update handler must not be called with invalid body")
	}
}

func TestWorkflowHandler_Update_InvalidStatus(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	body := validUpdateWorkflowBody()
	body["status"] = "invalid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with invalid status")
	}
}

func TestWorkflowHandler_Update_HandlerError_NotFound(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: errors.New("workflow not found")}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_Update_HandlerError_InvalidStatus(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: errors.New("invalid status")}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Update_HandlerError_UseDelete(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: errors.New("use delete to mark a workflow as deleted")}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Update_ScheduleError(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: domainworkflow.ErrScheduleIntervalTooShort}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Update_HandlerError_Internal(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Update_ReloadFailure(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view, nil},
		errs:  []error{nil, errors.New("database unavailable")},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if !update.called {
		t.Fatal("expected update handler to be called")
	}
}

func TestWorkflowHandler_Update_GetExisting_NotFound(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if update.called {
		t.Fatal("update handler must not be called when workflow does not exist")
	}
}

func TestWorkflowHandler_Activate_Success(t *testing.T) {
	w := sampleWorkflowEntity()
	w.Status = domainworkflow.StatusActive
	activate := &mockActivateWorkflowHandler{result: w}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", activeProject(), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !activate.called {
		t.Fatal("expected activate handler to be called")
	}
	if activate.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", activate.cmd.ProjectID)
	}
}

func TestWorkflowHandler_Activate_Unauthorized(t *testing.T) {
	activate := &mockActivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if activate.called {
		t.Fatal("activate handler must not be called without user")
	}
}

func TestWorkflowHandler_Activate_InvalidID(t *testing.T) {
	activate := &mockActivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", activeProject(), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/bad-id/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if activate.called {
		t.Fatal("activate handler must not be called with invalid id")
	}
}

func TestWorkflowHandler_Activate_NotFound(t *testing.T) {
	activate := &mockActivateWorkflowHandler{err: errors.New("workflow not found")}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", activeProject(), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_Activate_InvalidStatusTransition(t *testing.T) {
	activate := &mockActivateWorkflowHandler{err: domainworkflow.ErrInvalidStatusTransition}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", activeProject(), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Activate_HandlerError_Internal(t *testing.T) {
	activate := &mockActivateWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", activeProject(), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Deactivate_Success(t *testing.T) {
	w := sampleWorkflowEntity()
	deactivate := &mockDeactivateWorkflowHandler{result: w}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", activeProject(), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !deactivate.called {
		t.Fatal("expected deactivate handler to be called")
	}
}

func TestWorkflowHandler_Deactivate_Unauthorized(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if deactivate.called {
		t.Fatal("deactivate handler must not be called without user")
	}
}

func TestWorkflowHandler_Deactivate_InvalidID(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", activeProject(), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/bad-id/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deactivate.called {
		t.Fatal("deactivate handler must not be called with invalid id")
	}
}

func TestWorkflowHandler_Deactivate_NotFound(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{err: errors.New("workflow not found")}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", activeProject(), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_Deactivate_InvalidStatusTransition(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{err: domainworkflow.ErrInvalidStatusTransition}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", activeProject(), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Deactivate_HandlerError_Internal(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", activeProject(), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Delete_Success(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	deleteH := &mockDeleteWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
	if deleteH.cmd.ID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", deleteH.cmd.ID)
	}
}

func TestWorkflowHandler_Delete_Unauthorized(t *testing.T) {
	deleteH := &mockDeleteWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if deleteH.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called without user")
	}
}

func TestWorkflowHandler_Delete_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	deleteH := &mockDeleteWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called for wrong project")
	}
}

func TestWorkflowHandler_Delete_HandlerError_Internal(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	deleteH := &mockDeleteWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Delete_MissingActiveProject(t *testing.T) {
	deleteH := &mockDeleteWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called without active project")
	}
}

func TestWorkflowHandler_Delete_InvalidID(t *testing.T) {
	deleteH := &mockDeleteWorkflowHandler{}
	getByID := &mockGetWorkflowByIDHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/bad-id", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called || getByID.calls > 0 {
		t.Fatal("handlers must not be called with invalid id")
	}
}

func TestWorkflowHandler_Delete_GetExisting_NotFound(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("workflow not found")},
	}
	deleteH := &mockDeleteWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called when workflow does not exist")
	}
}

func TestWorkflowHandler_Delete_GetExisting_InternalError(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	deleteH := &mockDeleteWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, nil, deleteH, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/workflows/:id", activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/workflows/"+testutil.TestWorkflowID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called when get existing fails")
	}
}

func TestWorkflowHandler_Update_InvalidData(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called on validation error")
	}
}

func TestWorkflowHandler_Activate_MissingActiveProject(t *testing.T) {
	activate := &mockActivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, activate, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/activate", testutil.WithUserWithoutProject(testutil.TestUserID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if activate.called {
		t.Fatal("activate handler must not be called without active project")
	}
}

func TestWorkflowHandler_Deactivate_MissingActiveProject(t *testing.T) {
	deactivate := &mockDeactivateWorkflowHandler{}
	h := newWorkflowHandler(nil, nil, nil, deactivate, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows/:id/deactivate", testutil.WithUserWithoutProject(testutil.TestUserID), h.Deactivate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/"+testutil.TestWorkflowID.String()+"/deactivate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deactivate.called {
		t.Fatal("deactivate handler must not be called without active project")
	}
}

func TestWorkflowHandler_Update_GetExisting_InternalError(t *testing.T) {
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	update := &mockUpdateWorkflowHandler{}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if update.called {
		t.Fatal("update handler must not be called when get existing fails")
	}
}

func TestWorkflowHandler_Update_ScheduleTimezoneError(t *testing.T) {
	view := sampleWorkflowView()
	getByID := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateWorkflowHandler{err: domainworkflow.ErrInvalidScheduleTimezone}
	h := newWorkflowHandler(nil, update, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/workflows/:id", activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/workflows/"+testutil.TestWorkflowID.String(), validUpdateWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Create_WithOptionalFields(t *testing.T) {
	create := &mockCreateWorkflowHandler{result: sampleWorkflowEntity()}
	h := newWorkflowHandler(create, nil, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/workflows", activeProject(), h.Create)

	body := validCreateWorkflowBody()
	body["concurrency"] = 5
	body["notificationsEnabled"] = false

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows", body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if create.cmd.Concurrency != 5 {
		t.Fatalf("concurrency: got %d want 5", create.cmd.Concurrency)
	}
	if create.cmd.NotificationsEnabled {
		t.Fatal("expected notificationsEnabled false")
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
