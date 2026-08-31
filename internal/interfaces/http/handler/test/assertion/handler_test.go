package assertiontest

import (
	"errors"
	"net/http"
	"testing"

	domainassertion "go-api/internal/domain/assertion"
	domainworkflow "go-api/internal/domain/workflow"
	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestAssertionHandler_Create_Success(t *testing.T) {
	create := &mockCreateAssertionHandler{result: sampleAssertionEntity()}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}

	var out presenter.AssertionResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestAssertionID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestAssertionHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAssertionHandler_Create_NoActiveProject(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_InvalidWorkflowID(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPost,
		"/workflows/bad/steps/"+testutil.TestStepID.String()+"/assertions",
		validCreateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_InvalidStepID(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPost,
		"/workflows/"+testutil.TestWorkflowID.String()+"/steps/bad/assertions",
		validCreateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_WrongProject(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if create.called {
		t.Fatal("create handler must not be called for wrong project")
	}
}

func TestAssertionHandler_Create_WorkflowNotFound(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("workflow not found")},
	}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Create_InvalidSource(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	body := validCreateAssertionBody()
	body["source"] = "invalid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_InvalidOperator(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	body := validCreateAssertionBody()
	body["operator"] = "invalid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_StepNotFound(t *testing.T) {
	create := &mockCreateAssertionHandler{err: errors.New("step not found")}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Create_HandlerBadRequest(t *testing.T) {
	create := &mockCreateAssertionHandler{err: errors.New("path is required")}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validNotNullAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_QuotaExceeded(t *testing.T) {
	create := &mockCreateAssertionHandler{err: cmdquota.ErrAssertionQuotaExceeded}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validNotNullAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestAssertionHandler_ListByStep_Success(t *testing.T) {
	list := &mockListAssertionsByStepHandler{views: []domainassertion.AssertionView{*sampleAssertionView()}}
	h := newAssertionHandler(assertionMocks{
		listByStep:  list,
		getWorkflow: workflowMocksOK(),
		getStep:     stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(assertionsRoute(), activeProject(), h.ListByStep)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
}

func TestAssertionHandler_ListByStep_WrongProject(t *testing.T) {
	list := &mockListAssertionsByStepHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{listByStep: list, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Get(assertionsRoute(), activeProject(), h.ListByStep)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionsBasePath(), nil))
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

func TestAssertionHandler_ListByStep_StepNotFound(t *testing.T) {
	list := &mockListAssertionsByStepHandler{}
	getStep := &mockGetStepByIDHandler{
		errs: []error{errors.New("step not found")},
	}
	h := newAssertionHandler(assertionMocks{
		listByStep:  list,
		getWorkflow: workflowMocksOK(),
		getStep:     getStep,
	})

	app := testutil.NewTestApp()
	app.Get(assertionsRoute(), activeProject(), h.ListByStep)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_ListByStep_InternalError(t *testing.T) {
	list := &mockListAssertionsByStepHandler{err: errors.New("database unavailable")}
	h := newAssertionHandler(assertionMocks{
		listByStep:  list,
		getWorkflow: workflowMocksOK(),
		getStep:     stepMocksOK(),
	})

	app := testutil.NewTestApp()
	app.Get(assertionsRoute(), activeProject(), h.ListByStep)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionsBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestAssertionHandler_GetByID_Success(t *testing.T) {
	getByID := &mockGetAssertionByIDHandler{view: sampleAssertionView()}
	h := newAssertionHandler(assertionMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(assertionItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestAssertionHandler_GetByID_WrongProject(t *testing.T) {
	getByID := &mockGetAssertionByIDHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{getByID: getByID, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Get(assertionItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_GetByID_NotFound(t *testing.T) {
	getByID := &mockGetAssertionByIDHandler{err: domainassertion.ErrNotFound}
	h := newAssertionHandler(assertionMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(assertionItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_GetByID_InternalError(t *testing.T) {
	getByID := &mockGetAssertionByIDHandler{err: errors.New("database unavailable")}
	h := newAssertionHandler(assertionMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(assertionItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestAssertionHandler_Update_Success(t *testing.T) {
	update := &mockUpdateAssertionHandler{result: sampleAssertionEntity()}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		validUpdateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestAssertionHandler_Update_WrongProject(t *testing.T) {
	update := &mockUpdateAssertionHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		validUpdateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Update_NotFound(t *testing.T) {
	update := &mockUpdateAssertionHandler{err: domainassertion.ErrNotFound}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		validUpdateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Update_InvalidSource(t *testing.T) {
	update := &mockUpdateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	body := validUpdateAssertionBody()
	body["source"] = "invalid"

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		body,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteAssertionHandler{}
	h := newAssertionHandler(assertionMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(assertionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestAssertionHandler_Delete_WrongProject(t *testing.T) {
	deleteH := &mockDeleteAssertionHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newAssertionHandler(assertionMocks{deleteH: deleteH, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Delete(assertionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Delete_NotFound(t *testing.T) {
	deleteH := &mockDeleteAssertionHandler{err: domainassertion.ErrNotFound}
	h := newAssertionHandler(assertionMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(assertionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestAssertionHandler_Delete_InternalError(t *testing.T) {
	deleteH := &mockDeleteAssertionHandler{err: errors.New("database unavailable")}
	h := newAssertionHandler(assertionMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(assertionItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, assertionItemPath(testutil.TestAssertionID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestAssertionHandler_Create_GetWorkflowInternalError(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("database unavailable")},
	}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validCreateAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestAssertionHandler_Create_InvalidBody(t *testing.T) {
	create := &mockCreateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), map[string]any{"source": ""}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Update_Unauthorized(t *testing.T) {
	update := &mockUpdateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{update: update})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		validUpdateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestAssertionHandler_Update_InvalidBody(t *testing.T) {
	update := &mockUpdateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		map[string]any{"source": ""},
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Update_HandlerBadRequest(t *testing.T) {
	update := &mockUpdateAssertionHandler{err: errors.New("path is required")}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		validUpdateAssertionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Update_InvalidOperator(t *testing.T) {
	update := &mockUpdateAssertionHandler{}
	h := newAssertionHandler(assertionMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(assertionItemRoute(), activeProject(), h.Update)

	body := validUpdateAssertionBody()
	body["operator"] = "invalid"

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		assertionItemPath(testutil.TestAssertionID),
		body,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAssertionHandler_Create_ExpectedValueNil(t *testing.T) {
	create := &mockCreateAssertionHandler{result: sampleAssertionEntity()}
	h := newAssertionHandler(assertionMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(assertionsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, assertionsBasePath(), validNotNullAssertionBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if create.cmd.ExpectedValue != "" {
		t.Fatalf("expected value: got %q want empty", create.cmd.ExpectedValue)
	}
}
