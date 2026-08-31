package activitylogtest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	queryactivitylog "go-api/internal/application/query/activitylog"
	queryworkflow "go-api/internal/application/query/workflow"
	domainactivitylog "go-api/internal/domain/activitylog"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

type mockListByWorkflowHandler struct {
	called bool
	query  queryactivitylog.ListByWorkflowQuery
	views  []domainactivitylog.View
	total  int64
	err    error
}

func (m *mockListByWorkflowHandler) Handle(
	_ context.Context,
	q queryactivitylog.ListByWorkflowQuery,
) ([]domainactivitylog.View, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

type mockGetWorkflowByIDHandler struct {
	calls  int
	views  []*domainworkflow.WorkflowView
	errs   []error
}

func (m *mockGetWorkflowByIDHandler) Handle(
	_ context.Context,
	q queryworkflow.GetWorkflowByIDQuery,
) (*domainworkflow.WorkflowView, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get workflow call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if q.ID != testutil.TestWorkflowID {
		return nil, errors.New("unexpected workflow id")
	}
	return m.views[idx], nil
}

func newActivityLogHandler(
	list *mockListByWorkflowHandler,
	getWorkflow *mockGetWorkflowByIDHandler,
) *handler.ActivityLogHandler {
	if list == nil {
		list = &mockListByWorkflowHandler{}
	}
	if getWorkflow == nil {
		getWorkflow = &mockGetWorkflowByIDHandler{}
	}
	return handler.NewActivityLogHandler(list, getWorkflow)
}

func sampleWorkflowView() *domainworkflow.WorkflowView {
	return &domainworkflow.WorkflowView{
		ID:          testutil.TestWorkflowID,
		Name:        "Daily sync",
		Status:      domainworkflow.StatusActive,
		ProjectID:   testutil.TestProjectID,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func sampleActivityLogView() domainactivitylog.View {
	workflowID := testutil.TestWorkflowID
	return domainactivitylog.View{
		ID:              uuid.MustParse("01960000-0000-7000-8000-00000000000c"),
		ProjectID:       testutil.TestProjectID,
		Action:          "workflow.run.started",
		SubjectType:     "workflow_run",
		SubjectID:       testutil.TestWorkflowRunID,
		WorkflowID:      &workflowID,
		ActorType:       "system",
		Level:           "info",
		Message:         "Workflow run started",
		SourceEventID:   uuid.MustParse("01960000-0000-7000-8000-00000000000d"),
		SourceEventType: "workflow.run.started.v1",
		OccurredAt:      time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
		CreatedAt:       time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
	}
}

func workflowListPath() string {
	return "/workflows/" + testutil.TestWorkflowID.String() + "/activity-logs"
}

func TestActivityLogHandler_ListByWorkflow_Success(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	list := &mockListByWorkflowHandler{
		views: []domainactivitylog.View{sampleActivityLogView()},
		total: 1,
	}
	h := newActivityLogHandler(list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/activity-logs", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByWorkflow)

	req, err := testutil.JSONRequest(http.MethodGet, workflowListPath()+"?page=1&limit=10", nil)
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
	if getWorkflow.calls != 1 {
		t.Fatalf("get workflow calls: got %d want 1", getWorkflow.calls)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
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
		Members []presenter.ActivityLogResponse `json:"members"`
		Total   int                             `json:"total"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if len(out.Members) != 1 {
		t.Fatalf("members length: got %d want 1", len(out.Members))
	}
	if out.Total != 1 {
		t.Fatalf("total: got %d want 1", out.Total)
	}
}

func TestActivityLogHandler_ListByWorkflow_Unauthorized(t *testing.T) {
	list := &mockListByWorkflowHandler{}
	h := newActivityLogHandler(list, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/activity-logs", h.ListByWorkflow)

	req, err := testutil.JSONRequest(http.MethodGet, workflowListPath(), nil)
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
	if list.called {
		t.Fatal("list handler must not be called without user")
	}
}

func TestActivityLogHandler_ListByWorkflow_MissingActiveProject(t *testing.T) {
	list := &mockListByWorkflowHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{}
	h := newActivityLogHandler(list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/activity-logs", testutil.WithUserWithoutProject(testutil.TestUserID), h.ListByWorkflow)

	req, err := testutil.JSONRequest(http.MethodGet, workflowListPath(), nil)
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
	if getWorkflow.calls != 0 || list.called {
		t.Fatal("handlers must not be called without active project")
	}
}

func TestActivityLogHandler_ListByWorkflow_InvalidID(t *testing.T) {
	list := &mockListByWorkflowHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{}
	h := newActivityLogHandler(list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/activity-logs", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByWorkflow)

	req, err := testutil.JSONRequest(http.MethodGet, "/workflows/not-a-uuid/activity-logs", nil)
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
	if getWorkflow.calls != 0 || list.called {
		t.Fatal("handlers must not be called with invalid workflow id")
	}
}

func TestActivityLogHandler_ListByWorkflow_HandlerError_Business(t *testing.T) {
	tests := []struct {
		name      string
		workflow  *domainworkflow.WorkflowView
		getErr    error
		wantStatus int
	}{
		{
			name:       "workflow not found",
			workflow:   nil,
			getErr:     errors.New("workflow not found"),
			wantStatus: http.StatusNotFound,
		},
		{
			name: "wrong project",
			workflow: func() *domainworkflow.WorkflowView {
				view := sampleWorkflowView()
				view.ProjectID = uuid.New()
				return view
			}(),
			getErr:     nil,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getWorkflow := &mockGetWorkflowByIDHandler{
				views: []*domainworkflow.WorkflowView{tc.workflow},
				errs:  []error{tc.getErr},
			}
			list := &mockListByWorkflowHandler{}
			h := newActivityLogHandler(list, getWorkflow)

			app := testutil.NewTestApp()
			app.Get("/workflows/:workflowId/activity-logs", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByWorkflow)

			req, err := testutil.JSONRequest(http.MethodGet, workflowListPath(), nil)
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
			if list.called {
				t.Fatal("list handler must not be called when workflow is unavailable")
			}
		})
	}
}

func TestActivityLogHandler_ListByWorkflow_HandlerError_Internal(t *testing.T) {
	tests := []struct {
		name       string
		getViews   []*domainworkflow.WorkflowView
		getErrs    []error
		listErr    error
		wantCalled bool
	}{
		{
			name:     "get workflow fails",
			getViews: []*domainworkflow.WorkflowView{nil},
			getErrs:  []error{errors.New("database unavailable")},
		},
		{
			name:       "list fails",
			getViews:   []*domainworkflow.WorkflowView{sampleWorkflowView()},
			getErrs:    []error{nil},
			listErr:    errors.New("database unavailable"),
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			getWorkflow := &mockGetWorkflowByIDHandler{views: tc.getViews, errs: tc.getErrs}
			list := &mockListByWorkflowHandler{err: tc.listErr}
			h := newActivityLogHandler(list, getWorkflow)

			app := testutil.NewTestApp()
			app.Get("/workflows/:workflowId/activity-logs", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByWorkflow)

			req, err := testutil.JSONRequest(http.MethodGet, workflowListPath(), nil)
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
			if tc.wantCalled && !list.called {
				t.Fatal("expected list handler to be called")
			}
		})
	}
}

func TestActivityLogHandler_ListByWorkflow_InvalidQuery(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	list := &mockListByWorkflowHandler{}
	h := newActivityLogHandler(list, getWorkflow)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/activity-logs", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.ListByWorkflow)

	req, err := testutil.JSONRequest(http.MethodGet, workflowListPath()+"?page=not-a-number", nil)
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
