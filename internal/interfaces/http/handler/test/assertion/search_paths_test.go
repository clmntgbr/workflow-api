package assertiontest

import (
	"errors"
	"net/http"
	"testing"

	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/testutil"
)

func TestAssertionHandler_SearchPaths_Success(t *testing.T) {
	searchPaths := &mockSearchAssertionPathsHandler{
		paths: []string{"$.id", "$.status"},
		total: 2,
	}
	h := newAssertionHandler(assertionMocks{
		searchPaths: searchPaths,
		getWorkflow: workflowMocksOK(),
		getStep:     stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		assertionPathsPath()+"?page=1&limit=10",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !searchPaths.called {
		t.Fatal("expected search paths handler to be called")
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["total"] != float64(2) {
		t.Fatalf("total: got %#v", body["total"])
	}
}

func TestAssertionHandler_SearchPaths_InvalidWorkflowID(t *testing.T) {
	h := newAssertionHandler(assertionMocks{})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		"/workflows/bad/steps/"+testutil.TestStepID.String()+"/assertion-paths",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_SearchPaths_InvalidStepID(t *testing.T) {
	h := newAssertionHandler(assertionMocks{})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		"/workflows/"+testutil.TestWorkflowID.String()+"/steps/bad/assertion-paths",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_SearchPaths_WrongProject(t *testing.T) {
	searchPaths := &mockSearchAssertionPathsHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{
		searchPaths: searchPaths,
		getWorkflow: getWorkflow,
	})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionPathsPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if searchPaths.called {
		t.Fatal("search paths handler must not be called for wrong project")
	}
}

func TestAssertionHandler_SearchPaths_StepNotFound(t *testing.T) {
	searchPaths := &mockSearchAssertionPathsHandler{}
	getStep := &mockGetStepByIDHandler{
		errs: []error{errors.New("step not found")},
	}
	h := newAssertionHandler(assertionMocks{
		searchPaths: searchPaths,
		getWorkflow: workflowMocksOK(),
		getStep:     getStep,
	})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionPathsPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_SearchPaths_InvalidQuery(t *testing.T) {
	h := newAssertionHandler(assertionMocks{
		getWorkflow: workflowMocksOK(),
		getStep:     stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		assertionPathsPath()+"?page=abc",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_SearchPaths_InternalError(t *testing.T) {
	searchPaths := &mockSearchAssertionPathsHandler{err: errors.New("database unavailable")}
	h := newAssertionHandler(assertionMocks{
		searchPaths: searchPaths,
		getWorkflow: workflowMocksOK(),
		getStep:     stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(assertionPathsRoute(), activeProject(), h.SearchPaths)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionPathsPath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
