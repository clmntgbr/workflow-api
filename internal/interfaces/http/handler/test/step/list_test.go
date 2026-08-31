package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

func TestStepHandler_ListByWorkflow_Success(t *testing.T) {
	view := *sampleHTTPStepView()
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	list := &mockListStepsByWorkflowHandler{views: []domainstep.StepView{view}}
	latest := &mockLatestStepRunStatusHandler{
		result: map[uuid.UUID]domainsteprun.Status{
			testutil.TestStepID: domainsteprun.StatusSuccess,
		},
	}
	h := newStepHandler(stepMocks{
		getWorkflow:     getWorkflow,
		listByWorkflow:  list,
		latestRunStatus: latest,
	})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if !latest.called {
		t.Fatal("expected latest run status handler to be called")
	}
}

func TestStepHandler_ListByWorkflow_NoActiveProject(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_ListByWorkflow_InvalidWorkflowID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/bad/steps", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_ListByWorkflow_WorkflowNotFound(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("workflow not found")},
	}
	list := &mockListStepsByWorkflowHandler{}
	h := newStepHandler(stepMocks{getWorkflow: getWorkflow, listByWorkflow: list})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
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

func TestStepHandler_ListByWorkflow_WrongProject(t *testing.T) {
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	list := &mockListStepsByWorkflowHandler{}
	h := newStepHandler(stepMocks{getWorkflow: getWorkflow, listByWorkflow: list})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
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

func TestStepHandler_ListByWorkflow_GetWorkflowInternalError(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("database unavailable")},
	}
	h := newStepHandler(stepMocks{getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestStepHandler_ListByWorkflow_ListError(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	list := &mockListStepsByWorkflowHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{getWorkflow: getWorkflow, listByWorkflow: list})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestStepHandler_ListByWorkflow_StatusError(t *testing.T) {
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{sampleWorkflowView()},
		errs:  []error{nil},
	}
	list := &mockListStepsByWorkflowHandler{views: []domainstep.StepView{*sampleHTTPStepView()}}
	latest := &mockLatestStepRunStatusHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{
		getWorkflow:     getWorkflow,
		listByWorkflow:  list,
		latestRunStatus: latest,
	})

	app := testutil.NewTestApp()
	app.Get(stepsRoute(), activeProject(), h.ListByWorkflow)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
