package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_CreateDelay_Success(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{result: sampleDelayStepEntity()}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateDelayStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !createDelay.called {
		t.Fatal("expected create delay handler to be called")
	}
	if createDelay.cmd.DelayDurationSeconds != 30 {
		t.Fatalf("delay seconds: got %d want 30", createDelay.cmd.DelayDurationSeconds)
	}
}

func TestStepHandler_CreateDelay_MissingDuration(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := map[string]any{
		"type":     "delay",
		"name":     "Wait",
		"position": map[string]any{"x": 0.0, "y": 0.0},
	}

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if createDelay.called {
		t.Fatal("create delay handler must not be called without duration")
	}
}

func TestStepHandler_CreateDelay_InvalidDuration(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := validCreateDelayStepBody()
	body["delayDurationSeconds"] = 0

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if createDelay.called {
		t.Fatal("create delay handler must not be called with invalid duration")
	}
}

func TestStepHandler_CreateDelay_InvalidConfig(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{err: domainstep.ErrInvalidStepTypeConfig}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateDelayStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_CreateDelay_WorkflowNotFound(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{err: errors.New("workflow not found")}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateDelayStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_CreateDelay_QuotaExceeded(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{err: quotaErr()}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateDelayStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestStepHandler_CreateDelay_InternalError(t *testing.T) {
	createDelay := &mockCreateDelayStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{createDelay: createDelay})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateDelayStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
