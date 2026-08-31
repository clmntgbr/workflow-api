package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_UpdateDelay_Success(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleDelayStepView()},
		errs:  []error{nil},
	}
	updateDelay := &mockUpdateDelayStepHandler{result: sampleDelayStepEntity()}
	h := newStepHandler(stepMocks{getByID: getByID, updateDelay: updateDelay})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateDelayStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !updateDelay.called {
		t.Fatal("expected update delay handler to be called")
	}
}

func TestStepHandler_UpdateDelay_InvalidBody(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleDelayStepView()},
		errs:  []error{nil},
	}
	updateDelay := &mockUpdateDelayStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, updateDelay: updateDelay})

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
	if updateDelay.called {
		t.Fatal("update delay handler must not be called on validation error")
	}
}

func TestStepHandler_UpdateDelay_InvalidConfig(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleDelayStepView()},
		errs:  []error{nil},
	}
	updateDelay := &mockUpdateDelayStepHandler{err: domainstep.ErrInvalidStepTypeConfig}
	h := newStepHandler(stepMocks{getByID: getByID, updateDelay: updateDelay})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateDelayStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdateDelay_StepNotFound(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleDelayStepView()},
		errs:  []error{nil},
	}
	updateDelay := &mockUpdateDelayStepHandler{err: errors.New("step not found")}
	h := newStepHandler(stepMocks{getByID: getByID, updateDelay: updateDelay})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateDelayStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_UpdateDelay_InternalError(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleDelayStepView()},
		errs:  []error{nil},
	}
	updateDelay := &mockUpdateDelayStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{getByID: getByID, updateDelay: updateDelay})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateDelayStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
