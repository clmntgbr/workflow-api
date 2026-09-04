package projecttest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	projectcmd "go-api/internal/application/command/project"
	cmdquota "go-api/internal/application/command/quota"
	usercmd "go-api/internal/application/command/user"
	queryproject "go-api/internal/application/query/project"
	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

var otherProjectID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockCreateProjectHandler struct {
	called bool
	cmd    projectcmd.CreateProjectCommand
	result *domainproject.Project
	err    error
}

func (m *mockCreateProjectHandler) Handle(
	_ context.Context,
	cmd projectcmd.CreateProjectCommand,
) (*domainproject.Project, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateProjectHandler struct {
	called bool
	cmd    projectcmd.UpdateProjectCommand
	err    error
}

func (m *mockUpdateProjectHandler) Handle(_ context.Context, cmd projectcmd.UpdateProjectCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockDeleteProjectHandler struct {
	called bool
	id     uuid.UUID
	err    error
}

func (m *mockDeleteProjectHandler) Handle(_ context.Context, id uuid.UUID) error {
	m.called = true
	m.id = id
	return m.err
}

type mockRemoveMemberHandler struct {
	called bool
	cmd    projectcmd.RemoveProjectMemberCommand
	err    error
}

func (m *mockRemoveMemberHandler) Handle(_ context.Context, cmd projectcmd.RemoveProjectMemberCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
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
		return nil, errors.New("unexpected get by id call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	return m.views[idx], nil
}

type mockListProjectsByUserHandler struct {
	called bool
	query  queryproject.ListProjectsByUserQuery
	views  []domainproject.ProjectView
	total  int64
	err    error
}

func (m *mockListProjectsByUserHandler) Handle(
	_ context.Context,
	q queryproject.ListProjectsByUserQuery,
) ([]domainproject.ProjectView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockSetActiveProjectHandler struct {
	called bool
	cmd    usercmd.SetActiveProjectCommand
	err    error
}

func (m *mockSetActiveProjectHandler) Handle(_ context.Context, cmd usercmd.SetActiveProjectCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

func newProjectHandler(
	create *mockCreateProjectHandler,
	update *mockUpdateProjectHandler,
	deleteH *mockDeleteProjectHandler,
	remove *mockRemoveMemberHandler,
	getByID *mockGetProjectByIDHandler,
	list *mockListProjectsByUserHandler,
	setActive *mockSetActiveProjectHandler,
) *handler.ProjectHandler {
	if create == nil {
		create = &mockCreateProjectHandler{}
	}
	if update == nil {
		update = &mockUpdateProjectHandler{}
	}
	if deleteH == nil {
		deleteH = &mockDeleteProjectHandler{}
	}
	if remove == nil {
		remove = &mockRemoveMemberHandler{}
	}
	if getByID == nil {
		getByID = &mockGetProjectByIDHandler{}
	}
	if list == nil {
		list = &mockListProjectsByUserHandler{}
	}
	if setActive == nil {
		setActive = &mockSetActiveProjectHandler{}
	}
	return handler.NewProjectHandler(create, update, deleteH, remove, getByID, list, setActive)
}

func sampleProjectEntity() *domainproject.Project {
	return &domainproject.Project{
		ID:        testutil.TestProjectID,
		Name:      "My Project",
		MemberIDs: []uuid.UUID{testutil.TestUserID},
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleProjectView() *domainproject.ProjectView {
	e := sampleProjectEntity()
	return &domainproject.ProjectView{
		ID:        e.ID,
		Name:      e.Name,
		MemberIDs: e.MemberIDs,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func validCreateProjectBody() map[string]any {
	return map[string]any{"name": "My Project"}
}

func validUpdateProjectBody() map[string]any {
	return map[string]any{"name": "Updated Project"}
}

func TestProjectHandler_List_Success(t *testing.T) {
	view := sampleProjectView()
	list := &mockListProjectsByUserHandler{views: []domainproject.ProjectView{*view}, total: 1}
	h := newProjectHandler(nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects?page=2&limit=10&search=my", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if list.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", list.query.UserID)
	}
	if list.query.Query.Page != 2 || list.query.Query.Limit != 10 {
		t.Fatalf("paginate: got page=%d limit=%d", list.query.Query.Page, list.query.Query.Limit)
	}
	if list.query.Query.Search != "my" {
		t.Fatalf("search: got %q want %q", list.query.Query.Search, "my")
	}

	var out struct {
		Members    []presenter.ProjectDetailResponse `json:"members"`
		Total      int                               `json:"total"`
		Page       int                               `json:"page"`
		Limit      int                               `json:"limit"`
		TotalPages int                               `json:"totalPages"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if len(out.Members) != 1 {
		t.Fatalf("members length: got %d want 1", len(out.Members))
	}
	if out.Members[0].ID != testutil.TestProjectID.String() {
		t.Fatalf("member id: got %s", out.Members[0].ID)
	}
	if out.Total != 1 || out.Page != 2 || out.Limit != 10 || out.TotalPages != 1 {
		t.Fatalf("envelope: got total=%d page=%d limit=%d totalPages=%d", out.Total, out.Page, out.Limit, out.TotalPages)
	}
}

func TestProjectHandler_List_NormalizesPaginationBeforeHandler(t *testing.T) {
	list := &mockListProjectsByUserHandler{}
	h := newProjectHandler(nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects?limit=5000&page=0", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if list.query.Query.Limit != paginate.DefaultLimit {
		t.Fatalf("limit: got %d want %d", list.query.Query.Limit, paginate.DefaultLimit)
	}
	if list.query.Query.Page != 1 {
		t.Fatalf("page: got %d want 1", list.query.Query.Page)
	}
	if list.query.Query.OrderBy != paginate.OrderByAsc {
		t.Fatalf("orderBy: got %q want %q", list.query.Query.OrderBy, paginate.OrderByAsc)
	}
}

func TestProjectHandler_List_Unauthorized(t *testing.T) {
	list := &mockListProjectsByUserHandler{}
	h := newProjectHandler(nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if list.called {
		t.Fatal("list handler must not be called without user")
	}
}

func TestProjectHandler_List_InvalidQuery(t *testing.T) {
	list := &mockListProjectsByUserHandler{}
	h := newProjectHandler(nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects?page=not-a-number", nil))
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

func TestProjectHandler_List_HandlerError(t *testing.T) {
	list := &mockListProjectsByUserHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(nil, nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_Create_Success(t *testing.T) {
	create := &mockCreateProjectHandler{result: sampleProjectEntity()}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects", validCreateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
	if create.cmd.CreatorUserID != testutil.TestUserID {
		t.Fatalf("creator user id: got %s", create.cmd.CreatorUserID)
	}
	if create.cmd.Name != "My Project" {
		t.Fatalf("name: got %q", create.cmd.Name)
	}

	var out presenter.ProjectDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestProjectID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
	if !out.IsActive {
		t.Fatal("expected newly created project to be active")
	}
}

func TestProjectHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateProjectHandler{}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects", validCreateProjectBody()))
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

func TestProjectHandler_Create_InvalidRequestBody(t *testing.T) {
	create := &mockCreateProjectHandler{}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := http.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString("not-json"))
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

func TestProjectHandler_Create_InvalidData(t *testing.T) {
	create := &mockCreateProjectHandler{}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects", map[string]any{}))
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
	if _, ok := errorsMap["name"]; !ok {
		t.Fatalf("expected name validation error, got %#v", errorsMap)
	}
}

func TestProjectHandler_Create_HandlerError(t *testing.T) {
	create := &mockCreateProjectHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects", validCreateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_Create_QuotaExceeded(t *testing.T) {
	create := &mockCreateProjectHandler{err: cmdquota.ErrProjectQuotaExceeded}
	h := newProjectHandler(create, nil, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/projects", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects", validCreateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestProjectHandler_GetByID_Success(t *testing.T) {
	view := sampleProjectView()
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{view},
		errs:  []error{nil},
	}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if getByID.calls != 1 {
		t.Fatalf("get by id calls: got %d want 1", getByID.calls)
	}

	var out presenter.ProjectDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestProjectID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
	if !out.IsActive {
		t.Fatal("expected project to be active")
	}
}

func TestProjectHandler_GetByID_Unauthorized(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil))
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

func TestProjectHandler_GetByID_InvalidID(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects/not-a-uuid", nil))
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

func TestProjectHandler_GetByID_NotFound(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{nil},
		errs:  []error{errors.New("project not found")},
	}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProjectHandler_GetByID_HandlerError_Internal(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/projects/"+testutil.TestProjectID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_Update_Success(t *testing.T) {
	view := sampleProjectView()
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateProjectHandler{}
	h := newProjectHandler(nil, update, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), validUpdateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !update.called {
		t.Fatal("expected update handler to be called")
	}
	if update.cmd.ID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", update.cmd.ID)
	}
	if update.cmd.Name != "Updated Project" {
		t.Fatalf("name: got %q", update.cmd.Name)
	}
}

func TestProjectHandler_Update_Unauthorized(t *testing.T) {
	update := &mockUpdateProjectHandler{}
	getByID := &mockGetProjectByIDHandler{}
	h := newProjectHandler(nil, update, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), validUpdateProjectBody()))
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

func TestProjectHandler_Update_InvalidID(t *testing.T) {
	update := &mockUpdateProjectHandler{}
	h := newProjectHandler(nil, update, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/bad-id", validUpdateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with invalid id")
	}
}

func TestProjectHandler_Update_InvalidRequestBody(t *testing.T) {
	update := &mockUpdateProjectHandler{}
	h := newProjectHandler(nil, update, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := http.NewRequest(http.MethodPut, "/projects/"+testutil.TestProjectID.String(), bytes.NewBufferString("not-json"))
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

func TestProjectHandler_Update_InvalidData(t *testing.T) {
	update := &mockUpdateProjectHandler{}
	h := newProjectHandler(nil, update, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), map[string]any{}))
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

func TestProjectHandler_Update_NotFound(t *testing.T) {
	update := &mockUpdateProjectHandler{err: errors.New("project not found")}
	h := newProjectHandler(nil, update, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), validUpdateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProjectHandler_Update_HandlerError_Internal(t *testing.T) {
	update := &mockUpdateProjectHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(nil, update, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), validUpdateProjectBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_Update_ReloadFailure(t *testing.T) {
	view := sampleProjectView()
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{view},
		errs:  []error{errors.New("database unavailable")},
	}
	update := &mockUpdateProjectHandler{}
	h := newProjectHandler(nil, update, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/projects/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, "/projects/"+testutil.TestProjectID.String(), validUpdateProjectBody()))
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

func TestProjectHandler_Activate_Success(t *testing.T) {
	view := sampleProjectView()
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{view},
		errs:  []error{nil},
	}
	setActive := &mockSetActiveProjectHandler{}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+testutil.TestProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !setActive.called {
		t.Fatal("expected set active handler to be called")
	}
	if setActive.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", setActive.cmd.UserID)
	}
	if setActive.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", setActive.cmd.ProjectID)
	}

	var out presenter.ProjectDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if !out.IsActive {
		t.Fatal("expected activated project to be active")
	}
}

func TestProjectHandler_Activate_Unauthorized(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{}
	h := newProjectHandler(nil, nil, nil, nil, nil, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+testutil.TestProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if setActive.called {
		t.Fatal("set active handler must not be called without user")
	}
}

func TestProjectHandler_Activate_InvalidID(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{}
	h := newProjectHandler(nil, nil, nil, nil, nil, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/bad-id/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if setActive.called {
		t.Fatal("set active handler must not be called with invalid id")
	}
}

func TestProjectHandler_Activate_NotFound(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{err: errors.New("project not found")}
	h := newProjectHandler(nil, nil, nil, nil, nil, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+testutil.TestProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProjectHandler_Activate_NotMember(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{err: errors.New("user is not a member of the project")}
	h := newProjectHandler(nil, nil, nil, nil, nil, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+otherProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestProjectHandler_Activate_HandlerError_Internal(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(nil, nil, nil, nil, nil, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+testutil.TestProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_Activate_ReloadFailure(t *testing.T) {
	getByID := &mockGetProjectByIDHandler{
		views: []*domainproject.ProjectView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	setActive := &mockSetActiveProjectHandler{}
	h := newProjectHandler(nil, nil, nil, nil, getByID, nil, setActive)

	app := testutil.NewTestApp()
	app.Post("/projects/:id/activate", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Activate)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/projects/"+testutil.TestProjectID.String()+"/activate", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if !setActive.called {
		t.Fatal("expected set active handler to be called")
	}
}

func TestProjectHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteProjectHandler{}
	h := newProjectHandler(nil, nil, deleteH, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id", h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/projects/"+testutil.TestProjectID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
	if deleteH.id != testutil.TestProjectID {
		t.Fatalf("project id: got %s", deleteH.id)
	}
}

func TestProjectHandler_Delete_InvalidID(t *testing.T) {
	deleteH := &mockDeleteProjectHandler{}
	h := newProjectHandler(nil, nil, deleteH, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id", h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/projects/not-a-uuid", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called with invalid id")
	}
}

func TestProjectHandler_Delete_HandlerError(t *testing.T) {
	deleteH := &mockDeleteProjectHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(nil, nil, deleteH, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id", h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/projects/"+testutil.TestProjectID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestProjectHandler_RemoveMember_Success(t *testing.T) {
	remove := &mockRemoveMemberHandler{}
	h := newProjectHandler(nil, nil, nil, remove, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id/members/:userId", h.RemoveMember)

	path := "/projects/" + testutil.TestProjectID.String() + "/members/" + testutil.TestUserID.String()
	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !remove.called {
		t.Fatal("expected remove member handler to be called")
	}
	if remove.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", remove.cmd.ProjectID)
	}
	if remove.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", remove.cmd.UserID)
	}
}

func TestProjectHandler_RemoveMember_InvalidProjectID(t *testing.T) {
	remove := &mockRemoveMemberHandler{}
	h := newProjectHandler(nil, nil, nil, remove, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id/members/:userId", h.RemoveMember)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/projects/bad-id/members/"+testutil.TestUserID.String(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if remove.called {
		t.Fatal("remove member handler must not be called with invalid project id")
	}
}

func TestProjectHandler_RemoveMember_InvalidUserID(t *testing.T) {
	remove := &mockRemoveMemberHandler{}
	h := newProjectHandler(nil, nil, nil, remove, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id/members/:userId", h.RemoveMember)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, "/projects/"+testutil.TestProjectID.String()+"/members/bad-id", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if remove.called {
		t.Fatal("remove member handler must not be called with invalid user id")
	}
}

func TestProjectHandler_RemoveMember_NotFound(t *testing.T) {
	remove := &mockRemoveMemberHandler{err: errors.New("project not found")}
	h := newProjectHandler(nil, nil, nil, remove, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id/members/:userId", h.RemoveMember)

	path := "/projects/" + testutil.TestProjectID.String() + "/members/" + testutil.TestUserID.String()
	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestProjectHandler_RemoveMember_HandlerError(t *testing.T) {
	remove := &mockRemoveMemberHandler{err: errors.New("database unavailable")}
	h := newProjectHandler(nil, nil, nil, remove, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/projects/:id/members/:userId", h.RemoveMember)

	path := "/projects/" + testutil.TestProjectID.String() + "/members/" + testutil.TestUserID.String()
	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, path, nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
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
