package usertest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	usercmd "go-api/internal/application/command/user"
	queryuser "go-api/internal/application/query/user"
	domainuser "go-api/internal/domain/user"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

type mockGetUserByIDHandler struct {
	calls  int
	views  []*domainuser.UserView
	errs   []error
}

func (m *mockGetUserByIDHandler) Handle(
	_ context.Context,
	q queryuser.GetUserByIDQuery,
) (*domainuser.UserView, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get user call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if q.ID != testutil.TestUserID {
		return nil, errors.New("unexpected user id")
	}
	return m.views[idx], nil
}

type mockSetActiveProjectHandler struct {
	called bool
	cmd    usercmd.SetActiveProjectCommand
	err    error
}

func (m *mockSetActiveProjectHandler) Handle(
	_ context.Context,
	cmd usercmd.SetActiveProjectCommand,
) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

func newUserHandler(
	getByID *mockGetUserByIDHandler,
	setActive *mockSetActiveProjectHandler,
) *handler.UserHandler {
	if getByID == nil {
		getByID = &mockGetUserByIDHandler{}
	}
	if setActive == nil {
		setActive = &mockSetActiveProjectHandler{}
	}
	return handler.NewUserHandler(getByID, setActive)
}

func sampleUserView() *domainuser.UserView {
	projectID := testutil.TestProjectID
	return &domainuser.UserView{
		ID:              testutil.TestUserID,
		ClerkID:         "clerk_123",
		FirstName:       "Jane",
		LastName:        "Doe",
		Email:           "jane@example.com",
		ActiveProjectID: &projectID,
		CreatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	getByID := &mockGetUserByIDHandler{
		views: []*domainuser.UserView{sampleUserView()},
		errs:  []error{nil},
	}
	h := newUserHandler(getByID, nil)

	app := testutil.NewTestApp()
	app.Get("/user", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetUser)

	req, err := testutil.JSONRequest(http.MethodGet, "/user", nil)
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
		t.Fatalf("get user calls: got %d want 1", getByID.calls)
	}

	var out presenter.UserDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestUserID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
	if out.Email != "jane@example.com" {
		t.Fatalf("response email: got %q", out.Email)
	}
}

func TestUserHandler_GetUser_Unauthorized(t *testing.T) {
	getByID := &mockGetUserByIDHandler{}
	h := newUserHandler(getByID, nil)

	app := testutil.NewTestApp()
	app.Get("/user", h.GetUser)

	req, err := testutil.JSONRequest(http.MethodGet, "/user", nil)
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
	if getByID.calls != 0 {
		t.Fatal("get user handler must not be called without auth")
	}
}

func TestUserHandler_GetUser_HandlerError_Internal(t *testing.T) {
	getByID := &mockGetUserByIDHandler{
		views: []*domainuser.UserView{nil},
		errs:  []error{errors.New("database unavailable")},
	}
	h := newUserHandler(getByID, nil)

	app := testutil.NewTestApp()
	app.Get("/user", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetUser)

	req, err := testutil.JSONRequest(http.MethodGet, "/user", nil)
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

func TestUserHandler_SetActiveProject_Success(t *testing.T) {
	view := sampleUserView()
	getByID := &mockGetUserByIDHandler{
		views: []*domainuser.UserView{view},
		errs:  []error{nil},
	}
	setActive := &mockSetActiveProjectHandler{}
	h := newUserHandler(getByID, setActive)

	app := testutil.NewTestApp()
	app.Put("/user/active-project", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.SetActiveProject)

	req, err := testutil.JSONRequest(http.MethodPut, "/user/active-project", map[string]any{
		"projectId": testutil.TestProjectID.String(),
	})
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
	if !setActive.called {
		t.Fatal("expected set active project handler to be called")
	}
	if setActive.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", setActive.cmd.UserID)
	}
	if setActive.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", setActive.cmd.ProjectID)
	}
	if getByID.calls != 1 {
		t.Fatalf("get user calls after set: got %d want 1", getByID.calls)
	}
}

func TestUserHandler_SetActiveProject_Unauthorized(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{}
	h := newUserHandler(nil, setActive)

	app := testutil.NewTestApp()
	app.Put("/user/active-project", h.SetActiveProject)

	req, err := testutil.JSONRequest(http.MethodPut, "/user/active-project", map[string]any{
		"projectId": testutil.TestProjectID.String(),
	})
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
	if setActive.called {
		t.Fatal("set active project handler must not be called without auth")
	}
}

func TestUserHandler_SetActiveProject_InvalidBody(t *testing.T) {
	setActive := &mockSetActiveProjectHandler{}
	h := newUserHandler(nil, setActive)

	app := testutil.NewTestApp()
	app.Put("/user/active-project", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.SetActiveProject)

	req, err := http.NewRequest(http.MethodPut, "/user/active-project", bytes.NewBufferString("not-json"))
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
	if setActive.called {
		t.Fatal("set active project handler must not be called with invalid body")
	}
}

func TestUserHandler_SetActiveProject_InvalidID(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
	}{
		{
			name: "missing project id",
			body: map[string]any{},
		},
		{
			name: "invalid uuid",
			body: map[string]any{"projectId": "not-a-uuid"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setActive := &mockSetActiveProjectHandler{}
			h := newUserHandler(nil, setActive)

			app := testutil.NewTestApp()
			app.Put("/user/active-project", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.SetActiveProject)

			req, err := testutil.JSONRequest(http.MethodPut, "/user/active-project", tc.body)
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
			if setActive.called {
				t.Fatal("set active project handler must not be called with invalid id")
			}
		})
	}
}

func TestUserHandler_SetActiveProject_HandlerError_Business(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "project not found",
			err:        errors.New("project not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "not a member",
			err:        errors.New("user is not a member of the project"),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			setActive := &mockSetActiveProjectHandler{err: tc.err}
			h := newUserHandler(nil, setActive)

			app := testutil.NewTestApp()
			app.Put("/user/active-project", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.SetActiveProject)

			req, err := testutil.JSONRequest(http.MethodPut, "/user/active-project", map[string]any{
				"projectId": testutil.TestProjectID.String(),
			})
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status: got %d want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}

func TestUserHandler_SetActiveProject_HandlerError_Internal(t *testing.T) {
	tests := []struct {
		name     string
		setErr   error
		getViews []*domainuser.UserView
		getErrs  []error
	}{
		{
			name:   "set active fails",
			setErr: errors.New("database unavailable"),
		},
		{
			name:     "reload fails",
			getViews: []*domainuser.UserView{nil},
			getErrs:  []error{errors.New("database unavailable")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getByID := &mockGetUserByIDHandler{views: tc.getViews, errs: tc.getErrs}
			setActive := &mockSetActiveProjectHandler{err: tc.setErr}
			h := newUserHandler(getByID, setActive)

			app := testutil.NewTestApp()
			app.Put("/user/active-project", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.SetActiveProject)

			req, err := testutil.JSONRequest(http.MethodPut, "/user/active-project", map[string]any{
				"projectId": testutil.TestProjectID.String(),
			})
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
		})
	}
}
