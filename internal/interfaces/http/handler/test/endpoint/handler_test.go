package endpointtest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	endpointcmd "go-api/internal/application/command/endpoint"
	cmdquota "go-api/internal/application/command/quota"
	queryendpoint "go-api/internal/application/query/endpoint"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/httpquery"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mockCreateEndpointHandler struct {
	called bool
	cmd    endpointcmd.CreateEndpointCommand
	result *domainendpoint.Endpoint
	err    error
}

func (m *mockCreateEndpointHandler) Handle(
	_ context.Context,
	cmd endpointcmd.CreateEndpointCommand,
) (*domainendpoint.Endpoint, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockUpdateEndpointHandler struct {
	called bool
	cmd    endpointcmd.UpdateEndpointCommand
	err    error
}

func (m *mockUpdateEndpointHandler) Handle(_ context.Context, cmd endpointcmd.UpdateEndpointCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockDeleteEndpointHandler struct {
	called bool
	cmd    endpointcmd.DeleteEndpointCommand
	err    error
}

func (m *mockDeleteEndpointHandler) Handle(_ context.Context, cmd endpointcmd.DeleteEndpointCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockGetEndpointByIDHandler struct {
	calls  int
	views  []*domainendpoint.EndpointView
	errs   []error
}

func (m *mockGetEndpointByIDHandler) Handle(
	_ context.Context,
	q queryendpoint.GetEndpointByIDQuery,
) (*domainendpoint.EndpointView, error) {
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

type mockListEndpointsByProjectHandler struct {
	called bool
	query  queryendpoint.ListEndpointsByProjectQuery
	views  []domainendpoint.EndpointView
	total  int64
	err    error
}

func (m *mockListEndpointsByProjectHandler) Handle(
	_ context.Context,
	q queryendpoint.ListEndpointsByProjectQuery,
) ([]domainendpoint.EndpointView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockImportEndpointHandler struct {
	called bool
	cmd    endpointcmd.ImportEndpointsFromOpenAPICommand
	result []domainendpoint.Endpoint
	err    error
}

func (m *mockImportEndpointHandler) Handle(
	_ context.Context,
	cmd endpointcmd.ImportEndpointsFromOpenAPICommand,
) ([]domainendpoint.Endpoint, error) {
	m.called = true
	m.cmd = cmd
	if m.result == nil {
		m.result = []domainendpoint.Endpoint{}
	}
	return m.result, m.err
}

func mountEndpointRoutes(app *fiber.App, h *handler.EndpointHandler) {
	app.Post("/endpoints", h.Create)
	app.Post("/endpoints/import", h.ImportFromOpenAPI)
	app.Get("/endpoints", h.ListByProject)
	app.Get("/endpoints/:id", h.GetByID)
	app.Put("/endpoints/:id", h.Update)
	app.Delete("/endpoints/:id", h.Delete)
}

func newEndpointHandler(
	create *mockCreateEndpointHandler,
	update *mockUpdateEndpointHandler,
	deleteH *mockDeleteEndpointHandler,
	getByID *mockGetEndpointByIDHandler,
	list *mockListEndpointsByProjectHandler,
	importH *mockImportEndpointHandler,
) *handler.EndpointHandler {
	if create == nil {
		create = &mockCreateEndpointHandler{}
	}
	if update == nil {
		update = &mockUpdateEndpointHandler{}
	}
	if deleteH == nil {
		deleteH = &mockDeleteEndpointHandler{}
	}
	if getByID == nil {
		getByID = &mockGetEndpointByIDHandler{}
	}
	if list == nil {
		list = &mockListEndpointsByProjectHandler{}
	}
	if importH == nil {
		importH = &mockImportEndpointHandler{err: errors.New("import not configured")}
	}
	return handler.NewEndpointHandler(
		create,
		importH,
		update,
		deleteH,
		getByID,
		list,
	)
}

func validCreateEndpointBody() map[string]any {
	return map[string]any{
		"name":           "Users API",
		"url":            "https://api.example.com/users",
		"method":         "GET",
		"timeout":        30000,
		"retryOnFailure": false,
		"retryCount":     0,
		"retryDelay":     10000,
	}
}

func validUpdateEndpointBody() map[string]any {
	body := validCreateEndpointBody()
	body["status"] = "active"
	return body
}

func sampleEndpointEntity() *domainendpoint.Endpoint {
	return &domainendpoint.Endpoint{
		ID:             testutil.TestEndpointID,
		Name:           "Users API",
		URL:            "https://api.example.com/users",
		Method:         domainendpoint.MethodGET,
		Headers:        map[string]string{},
		Query:          httpquery.Empty(),
		Body:           map[string]any{},
		Timeout:        30000,
		RetryOnFailure: false,
		RetryCount:     0,
		RetryDelay:     10000,
		Status:         domainendpoint.StatusActive,
		ProjectID:      testutil.TestProjectID,
		CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleEndpointView() *domainendpoint.EndpointView {
	e := sampleEndpointEntity()
	return &domainendpoint.EndpointView{
		ID:             e.ID,
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
		ProjectID:      e.ProjectID,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}

func TestEndpointHandler_Create_Success(t *testing.T) {
	create := &mockCreateEndpointHandler{result: sampleEndpointEntity()}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", validCreateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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
		t.Fatalf("user id: got %s want %s", create.cmd.UserID, testutil.TestUserID)
	}
	if create.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s want %s", create.cmd.ProjectID, testutil.TestProjectID)
	}
	if create.cmd.Name != "Users API" {
		t.Fatalf("name: got %q", create.cmd.Name)
	}
	if create.cmd.Method != domainendpoint.MethodGET {
		t.Fatalf("method: got %s want GET", create.cmd.Method)
	}

	var out presenter.EndpointDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestEndpointID.String() {
		t.Fatalf("response id: got %s want %s", out.ID, testutil.TestEndpointID)
	}
	if out.Name != "Users API" {
		t.Fatalf("response name: got %q", out.Name)
	}
}

func TestEndpointHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateEndpointHandler{}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", h.Create)

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", validCreateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Create_InvalidData(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		want string
	}{
		{
			name: "missing name",
			body: map[string]any{
				"url":            "https://api.example.com/users",
				"method":         "GET",
				"timeout":        30000,
				"retryOnFailure": false,
				"retryCount":     0,
				"retryDelay":     10000,
			},
			want: "name",
		},
		{
			name: "missing url",
			body: map[string]any{
				"name":           "Users API",
				"method":         "GET",
				"timeout":        30000,
				"retryOnFailure": false,
				"retryCount":     0,
				"retryDelay":     10000,
			},
			want: "url",
		},
		{
			name: "invalid url",
			body: map[string]any{
				"name":           "Users API",
				"url":            "not-a-url",
				"method":         "GET",
				"timeout":        30000,
				"retryOnFailure": false,
				"retryCount":     0,
				"retryDelay":     10000,
			},
			want: "url",
		},
		{
			name: "invalid method",
			body: map[string]any{
				"name":           "Users API",
				"url":            "https://api.example.com/users",
				"method":         "INVALID",
				"timeout":        30000,
				"retryOnFailure": false,
				"retryCount":     0,
				"retryDelay":     10000,
			},
			want: "method",
		},
		{
			name: "timeout too low",
			body: map[string]any{
				"name":           "Users API",
				"url":            "https://api.example.com/users",
				"method":         "GET",
				"timeout":        1000,
				"retryOnFailure": false,
				"retryCount":     0,
				"retryDelay":     10000,
			},
			want: "timeout",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			create := &mockCreateEndpointHandler{}
			h := newEndpointHandler(create, nil, nil, nil, nil, nil)

			app := testutil.NewTestApp()
			app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

			req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", tc.body)
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := app.Test(req)
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
				t.Fatalf("expected errors map in response, got %#v", body)
			}
			if _, ok := errorsMap[tc.want]; !ok {
				t.Fatalf("expected validation error for %q, got %#v", tc.want, errorsMap)
			}
		})
	}
}

func TestEndpointHandler_Create_HandlerError(t *testing.T) {
	create := &mockCreateEndpointHandler{err: errors.New("failed to create endpoint")}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", validCreateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
}

func TestEndpointHandler_Create_HandlerError_QuotaExceeded(t *testing.T) {
	create := &mockCreateEndpointHandler{err: cmdquota.ErrEndpointQuotaExceeded}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", validCreateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestEndpointHandler_GetByID_Success(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if getByID.calls != 1 {
		t.Fatalf("get by id calls: got %d want 1", getByID.calls)
	}

	var out presenter.EndpointDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestEndpointID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestEndpointHandler_GetByID_Unauthorized(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_GetByID_InvalidID(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/not-a-uuid", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_GetByID_HandlerError_NotFound(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{nil},
		errs:  []error{errors.New("endpoint not found")},
	}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestEndpointHandler_GetByID_HandlerError_WrongProject(t *testing.T) {
	view := sampleEndpointView()
	view.ProjectID = uuid.New()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestEndpointHandler_Update_Success(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view, view},
		errs:  []error{nil, nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !update.called {
		t.Fatal("expected update handler to be called")
	}
	if update.cmd.ID != testutil.TestEndpointID {
		t.Fatalf("endpoint id: got %s", update.cmd.ID)
	}
	if update.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", update.cmd.UserID)
	}
	if getByID.calls != 2 {
		t.Fatalf("get by id calls: got %d want 2", getByID.calls)
	}
}

func TestEndpointHandler_Update_Unauthorized(t *testing.T) {
	update := &mockUpdateEndpointHandler{}
	getByID := &mockGetEndpointByIDHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_InvalidID(t *testing.T) {
	update := &mockUpdateEndpointHandler{}
	getByID := &mockGetEndpointByIDHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/bad-id", validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_InvalidData(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	body := validUpdateEndpointBody()
	delete(body, "status")

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_HandlerError_NotFound(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{err: errors.New("endpoint not found")}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestEndpointHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
	if deleteH.cmd.ID != testutil.TestEndpointID {
		t.Fatalf("endpoint id: got %s", deleteH.cmd.ID)
	}
	if deleteH.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", deleteH.cmd.ProjectID)
	}
}

func TestEndpointHandler_Delete_Unauthorized(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Delete_InvalidID(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/not-a-uuid", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Delete_HandlerError_NotFound(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{err: errors.New("endpoint not found")}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestEndpointHandler_ListByProject_Success(t *testing.T) {
	view := sampleEndpointView()
	list := &mockListEndpointsByProjectHandler{
		views: []domainendpoint.EndpointView{*view},
		total: 1,
	}
	h := newEndpointHandler(nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByProject)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints?page=1&limit=10", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

	var out struct {
		Members []presenter.EndpointListResponse `json:"members"`
		Total   int                              `json:"total"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if len(out.Members) != 1 {
		t.Fatalf("members length: got %d want 1", len(out.Members))
	}
	if out.Total != 1 {
		t.Fatalf("total: got %d want 1", out.Total)
	}
}

func TestEndpointHandler_ListByProject_Unauthorized(t *testing.T) {
	list := &mockListEndpointsByProjectHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints", h.ListByProject)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_ListByProject_HandlerError_InvalidMethodFilter(t *testing.T) {
	list := &mockListEndpointsByProjectHandler{err: errors.New("invalid endpoint method: FOO")}
	h := newEndpointHandler(nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByProject)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints?method=FOO", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_ListByProject_HandlerError_Internal(t *testing.T) {
	list := &mockListEndpointsByProjectHandler{err: errors.New("database unavailable")}
	h := newEndpointHandler(nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByProject)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestEndpointHandler_Create_InvalidRequestBody(t *testing.T) {
	create := &mockCreateEndpointHandler{}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	req, err := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBufferString("not-json"))
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

func TestEndpointHandler_Create_MissingActiveProject(t *testing.T) {
	create := &mockCreateEndpointHandler{}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", validCreateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Create_InvalidURLQuery(t *testing.T) {
	create := &mockCreateEndpointHandler{}
	h := newEndpointHandler(create, nil, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Create)

	body := validCreateEndpointBody()
	body["url"] = "https://api.example.com/users?bad%"

	req, err := testutil.JSONRequest(http.MethodPost, "/endpoints", body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid url query")
	}
}

func TestEndpointHandler_GetByID_HandlerError_Internal(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newEndpointHandler(nil, nil, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetByID)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestEndpointHandler_Update_MissingActiveProject(t *testing.T) {
	update := &mockUpdateEndpointHandler{}
	getByID := &mockGetEndpointByIDHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_InvalidRequestBody(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := http.NewRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), bytes.NewBufferString("not-json"))
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

func TestEndpointHandler_Update_GetExisting_InternalError(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_HandlerError_WrongProject(t *testing.T) {
	view := sampleEndpointView()
	view.ProjectID = uuid.New()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_InvalidMethod(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	body := validUpdateEndpointBody()
	body["method"] = "INVALID"

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with invalid method")
	}
}

func TestEndpointHandler_Update_InvalidStatus(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	body := validUpdateEndpointBody()
	body["status"] = "deleted"

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with deleted status")
	}
}

func TestEndpointHandler_Update_InvalidURLQuery(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	body := validUpdateEndpointBody()
	body["url"] = "https://api.example.com/users?bad%"

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with invalid url query")
	}
}

func TestEndpointHandler_Update_HandlerError_InvalidStatus(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{err: errors.New("invalid status")}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_Update_HandlerError_Internal(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{err: errors.New("database unavailable")}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestEndpointHandler_Update_ReloadFailure(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view, nil},
		errs:  []error{nil, errors.New("database unavailable")},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Delete_MissingActiveProject(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", testutil.WithUserWithoutProject(testutil.TestUserID), h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Update_GetExisting_NotFound(t *testing.T) {
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{nil},
		errs:  []error{errors.New("endpoint not found")},
	}
	update := &mockUpdateEndpointHandler{}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if update.called {
		t.Fatal("update handler must not be called when endpoint does not exist")
	}
}

func TestEndpointHandler_Update_HandlerError_UseDelete(t *testing.T) {
	view := sampleEndpointView()
	getByID := &mockGetEndpointByIDHandler{
		views: []*domainendpoint.EndpointView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateEndpointHandler{err: errors.New("use delete to mark an endpoint as deleted")}
	h := newEndpointHandler(nil, update, nil, getByID, nil, nil)

	app := testutil.NewTestApp()
	app.Put("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Update)

	req, err := testutil.JSONRequest(http.MethodPut, "/endpoints/"+testutil.TestEndpointID.String(), validUpdateEndpointBody())
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_ListByProject_InvalidQuery(t *testing.T) {
	list := &mockListEndpointsByProjectHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, list, nil)

	app := testutil.NewTestApp()
	app.Get("/endpoints", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByProject)

	req, err := testutil.JSONRequest(http.MethodGet, "/endpoints?page=not-a-number", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_Delete_HandlerError_Internal(t *testing.T) {
	deleteH := &mockDeleteEndpointHandler{err: errors.New("database unavailable")}
	h := newEndpointHandler(nil, nil, deleteH, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Delete("/endpoints/:id", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.Delete)

	req, err := testutil.JSONRequest(http.MethodDelete, "/endpoints/"+testutil.TestEndpointID.String(), nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
