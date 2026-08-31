package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_UpdateHTTP_Success(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{result: sampleHTTPStepEntity()}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
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

	var out presenter.StepDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestStepID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestStepHandler_UpdateHTTP_Unauthorized(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if update.called {
		t.Fatal("update handler must not be called without user")
	}
}

func TestStepHandler_UpdateHTTP_NoActiveProject(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdateHTTP_InvalidWorkflowID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		"/workflows/bad/steps/"+testutil.TestStepID.String(),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdateHTTP_InvalidStepID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepsBasePath()+"/bad",
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdateHTTP_StepNotFound(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		errs: []error{errors.New("step not found")},
	}
	update := &mockUpdateStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
	if update.called {
		t.Fatal("update handler must not be called when step not found")
	}
}

func TestStepHandler_UpdateHTTP_WrongProject(t *testing.T) {
	view := sampleHTTPStepView()
	view.ProjectID = otherProjectID
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{view},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
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

func TestStepHandler_UpdateHTTP_GetInternalError(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		errs: []error{errors.New("database unavailable")},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestStepHandler_UpdateHTTP_InvalidBody(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		map[string]any{"name": ""},
	))
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

func TestStepHandler_UpdateHTTP_InvalidURL(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	body := validUpdateHTTPStepBody()
	body["url"] = "://bad-url"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPut, stepItemPath(testutil.TestStepID), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if update.called {
		t.Fatal("update handler must not be called with invalid url")
	}
}

func TestStepHandler_UpdateHTTP_UpdateNotFound(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{err: errors.New("step not found")}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_UpdateHTTP_UpdateInternalError(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	update := &mockUpdateStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{getByID: getByID, update: update})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateHTTPStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
