package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_UpdateCondition_Success(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleConditionStepView()},
		errs:  []error{nil},
	}
	updateCondition := &mockUpdateConditionStepHandler{result: sampleConditionStepEntity()}
	h := newStepHandler(stepMocks{getByID: getByID, updateCondition: updateCondition})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateConditionStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !updateCondition.called {
		t.Fatal("expected update condition handler to be called")
	}
}

func TestStepHandler_UpdateCondition_InvalidBody(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleConditionStepView()},
		errs:  []error{nil},
	}
	updateCondition := &mockUpdateConditionStepHandler{}
	h := newStepHandler(stepMocks{getByID: getByID, updateCondition: updateCondition})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		map[string]any{"name": "Check", "expression": ""},
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if updateCondition.called {
		t.Fatal("update condition handler must not be called on validation error")
	}
}

func TestStepHandler_UpdateCondition_InvalidConfig(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleConditionStepView()},
		errs:  []error{nil},
	}
	updateCondition := &mockUpdateConditionStepHandler{err: domainstep.ErrInvalidStepTypeConfig}
	h := newStepHandler(stepMocks{getByID: getByID, updateCondition: updateCondition})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateConditionStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdateCondition_StepNotFound(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleConditionStepView()},
		errs:  []error{nil},
	}
	updateCondition := &mockUpdateConditionStepHandler{err: errors.New("step not found")}
	h := newStepHandler(stepMocks{getByID: getByID, updateCondition: updateCondition})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateConditionStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_UpdateCondition_InternalError(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleConditionStepView()},
		errs:  []error{nil},
	}
	updateCondition := &mockUpdateConditionStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{getByID: getByID, updateCondition: updateCondition})

	app := testutil.NewTestApp()
	app.Put(stepItemRoute(), activeProject(), h.Update)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepItemPath(testutil.TestStepID),
		validUpdateConditionStepBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
