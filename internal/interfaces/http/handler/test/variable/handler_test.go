package variabletest

import (
	"errors"
	"net/http"
	"testing"

	variablecmd "go-api/internal/application/command/variable"
	cmdquota "go-api/internal/application/command/quota"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestVariableHandler_Create_Success(t *testing.T) {
	create := &mockCreateVariableHandler{result: sampleVariableEntity()}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}

	var out presenter.VariableResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestVariableID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestVariableHandler_Create_Unauthorized(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
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

func TestVariableHandler_Create_NoActiveProject(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_InvalidWorkflowID(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/bad/variables", validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_WorkflowNotFound(t *testing.T) {
	create := &mockCreateVariableHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("workflow not found")},
	}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if create.called {
		t.Fatal("create handler must not be called when workflow not found")
	}
}

func TestVariableHandler_Create_WrongProject(t *testing.T) {
	create := &mockCreateVariableHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
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

func TestVariableHandler_Create_GetWorkflowInternalError(t *testing.T) {
	create := &mockCreateVariableHandler{}
	getWorkflow := &mockGetWorkflowByIDHandler{
		errs: []error{errors.New("database unavailable")},
	}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestVariableHandler_Create_InvalidBody(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), map[string]any{"name": ""}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_InvalidStepID(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	body := validCreateVariableBody()
	body["stepId"] = "bad"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_InvalidKind(t *testing.T) {
	create := &mockCreateVariableHandler{}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	body := validCreateVariableBody()
	body["kind"] = "invalid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_DuplicateKey(t *testing.T) {
	create := &mockCreateVariableHandler{err: domainvariable.ErrDuplicateKey}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestVariableHandler_Create_StepNotFound(t *testing.T) {
	create := &mockCreateVariableHandler{err: errors.New("step not found")}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVariableHandler_Create_HandlerBadRequest(t *testing.T) {
	create := &mockCreateVariableHandler{err: errors.New("path is required")}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Create_QuotaExceeded(t *testing.T) {
	create := &mockCreateVariableHandler{err: cmdquota.ErrVariableQuotaExceeded}
	h := newVariableHandler(variableMocks{create: create, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Post(variablesRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, variablesBasePath(), validCreateVariableBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestVariableHandler_List_Success(t *testing.T) {
	list := &mockListVariablesByWorkflowHandler{views: []domainvariable.VariableView{*sampleVariableView()}}
	h := newVariableHandler(variableMocks{listByWorkflow: list, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variablesRoute(), activeProject(), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variablesBasePath(), nil))
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

func TestVariableHandler_List_WrongProject(t *testing.T) {
	list := &mockListVariablesByWorkflowHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{listByWorkflow: list, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Get(variablesRoute(), activeProject(), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variablesBasePath(), nil))
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

func TestVariableHandler_List_InternalError(t *testing.T) {
	list := &mockListVariablesByWorkflowHandler{err: errors.New("database unavailable")}
	h := newVariableHandler(variableMocks{listByWorkflow: list, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variablesRoute(), activeProject(), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variablesBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestVariableHandler_GetByID_Success(t *testing.T) {
	getByID := &mockGetVariableByIDHandler{view: sampleVariableView()}
	h := newVariableHandler(variableMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variableItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestVariableHandler_GetByID_WrongProject(t *testing.T) {
	getByID := &mockGetVariableByIDHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{getByID: getByID, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Get(variableItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if getByID.called {
		t.Fatal("get by id handler must not be called for wrong project")
	}
}

func TestVariableHandler_GetByID_NotFound(t *testing.T) {
	getByID := &mockGetVariableByIDHandler{err: domainvariable.ErrNotFound}
	h := newVariableHandler(variableMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variableItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVariableHandler_GetByID_InternalError(t *testing.T) {
	getByID := &mockGetVariableByIDHandler{err: errors.New("database unavailable")}
	h := newVariableHandler(variableMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variableItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestVariableHandler_Update_Success(t *testing.T) {
	update := &mockUpdateVariableHandler{result: sampleVariableEntity()}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !update.called {
		t.Fatal("expected update handler to be called")
	}
}

func TestVariableHandler_Update_WrongProject(t *testing.T) {
	update := &mockUpdateVariableHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
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

func TestVariableHandler_Update_NotFound(t *testing.T) {
	update := &mockUpdateVariableHandler{err: domainvariable.ErrNotFound}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVariableHandler_Update_DuplicateKey(t *testing.T) {
	update := &mockUpdateVariableHandler{err: domainvariable.ErrDuplicateKey}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}
}

func TestVariableHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteVariableHandler{}
	h := newVariableHandler(variableMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(variableItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
}

func TestVariableHandler_Delete_WrongProject(t *testing.T) {
	deleteH := &mockDeleteVariableHandler{}
	view := sampleWorkflowView()
	view.ProjectID = otherProjectID
	getWorkflow := &mockGetWorkflowByIDHandler{
		views: []*domainworkflow.WorkflowView{view},
		errs:  []error{nil},
	}
	h := newVariableHandler(variableMocks{deleteH: deleteH, getWorkflow: getWorkflow})

	app := testutil.NewTestApp()
	app.Delete(variableItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, variableItemPath(testutil.TestVariableID), nil))
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

func TestVariableHandler_Delete_NotFound(t *testing.T) {
	deleteH := &mockDeleteVariableHandler{err: domainvariable.ErrNotFound}
	h := newVariableHandler(variableMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(variableItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestVariableHandler_Delete_InUse(t *testing.T) {
	deleteH := &mockDeleteVariableHandler{
		err: &variablecmd.VariableInUseError{
			Steps: []domainstep.StepView{*sampleStepView()},
		},
	}
	h := newVariableHandler(variableMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(variableItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusConflict)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Variable is used by one or more steps" {
		t.Fatalf("message: got %#v", body["message"])
	}
}

func TestVariableHandler_Delete_InternalError(t *testing.T) {
	deleteH := &mockDeleteVariableHandler{err: errors.New("database unavailable")}
	h := newVariableHandler(variableMocks{deleteH: deleteH, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Delete(variableItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, variableItemPath(testutil.TestVariableID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestVariableHandler_Update_Unauthorized(t *testing.T) {
	update := &mockUpdateVariableHandler{}
	h := newVariableHandler(variableMocks{update: update})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
}

func TestVariableHandler_Update_InvalidBody(t *testing.T) {
	update := &mockUpdateVariableHandler{}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		map[string]any{"name": ""},
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_Update_HandlerBadRequest(t *testing.T) {
	update := &mockUpdateVariableHandler{err: errors.New("path is required")}
	h := newVariableHandler(variableMocks{update: update, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Put(variableItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		variableItemPath(testutil.TestVariableID),
		validUpdateVariableBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_List_NoActiveProject(t *testing.T) {
	list := &mockListVariablesByWorkflowHandler{}
	h := newVariableHandler(variableMocks{listByWorkflow: list})

	app := testutil.NewTestApp()
	app.Get(variablesRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.List)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variablesBasePath(), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestVariableHandler_GetByID_InvalidVariableID(t *testing.T) {
	getByID := &mockGetVariableByIDHandler{}
	h := newVariableHandler(variableMocks{getByID: getByID, getWorkflow: workflowMocksOK()})

	app := testutil.NewTestApp()
	app.Get(variableItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, variablesBasePath()+"/bad", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
