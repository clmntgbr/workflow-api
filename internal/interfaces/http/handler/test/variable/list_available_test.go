package variabletest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

func TestVariableHandler_ListAvailable_Success(t *testing.T) {
	listAvailable := &mockListAvailableVariablesHandler{
		views: []domainvariable.VariableView{*sampleVariableView()},
	}
	h := newVariableHandler(variableMocks{
		listAvailable: listAvailable,
		getWorkflow:   workflowMocksOK(),
		getStep:       stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		availableVariablesPath(testutil.TestStepID),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !listAvailable.called {
		t.Fatal("expected list available handler to be called")
	}
}

func TestVariableHandler_ListAvailable_InvalidWorkflowID(t *testing.T) {
	h := newVariableHandler(variableMocks{})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		"/workflows/bad/steps/"+testutil.TestStepID.String()+"/variables",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_ListAvailable_InvalidStepID(t *testing.T) {
	h := newVariableHandler(variableMocks{})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		"/workflows/"+testutil.TestWorkflowID.String()+"/steps/bad/variables",
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_ListAvailable_WrongProject(t *testing.T) {
	listAvailable := &mockListAvailableVariablesHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{
		listAvailable: listAvailable,
		getWorkflow:   getWorkflow,
	})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		availableVariablesPath(testutil.TestStepID),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if listAvailable.called {
		t.Fatal("list available handler must not be called for wrong project")
	}
}

func TestVariableHandler_ListAvailable_StepNotFound(t *testing.T) {
	listAvailable := &mockListAvailableVariablesHandler{}
	getStep := &mockGetStepByIDHandler{
		errs: []error{errors.New("step not found")},
	}
	h := newVariableHandler(variableMocks{
		listAvailable: listAvailable,
		getWorkflow:   workflowMocksOK(),
		getStep:       getStep,
	})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		availableVariablesPath(testutil.TestStepID),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if listAvailable.called {
		t.Fatal("list available handler must not be called when step not found")
	}
}

func TestVariableHandler_ListAvailable_WrongWorkflowStep(t *testing.T) {
	listAvailable := &mockListAvailableVariablesHandler{}
	view := sampleStepView()
	view.WorkflowID = uuid.New()
	getStep := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{
		listAvailable: listAvailable,
		getWorkflow:   workflowMocksOK(),
		getStep:       getStep,
	})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		availableVariablesPath(testutil.TestStepID),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVariableHandler_ListAvailable_InternalError(t *testing.T) {
	listAvailable := &mockListAvailableVariablesHandler{err: errors.New("database unavailable")}
	h := newVariableHandler(variableMocks{
		listAvailable: listAvailable,
		getWorkflow:   workflowMocksOK(),
		getStep:       stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(availableVariablesRoute(), activeProject(), h.ListAvailable)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		availableVariablesPath(testutil.TestStepID),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func stepMocksOK() *mockGetStepByIDHandler {
	return &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleStepView()},
		errs:  []error{nil},
	}
}
