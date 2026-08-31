package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_CreateCondition_Success(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{result: sampleConditionStepEntity()}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateConditionStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !createCondition.called {
		t.Fatal("expected create condition handler to be called")
	}
	if createCondition.cmd.Expression != "status == 200" {
		t.Fatalf("expression: got %q", createCondition.cmd.Expression)
	}
}

func TestStepHandler_CreateCondition_MissingExpression(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := map[string]any{
		"type":     "condition",
		"name":     "Check Status",
		"position": map[string]any{"x": 0.0, "y": 0.0},
	}

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if createCondition.called {
		t.Fatal("create condition handler must not be called without expression")
	}
}

func TestStepHandler_CreateCondition_EmptyExpression(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := validCreateConditionStepBody()
	body["expression"] = "   "

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if createCondition.called {
		t.Fatal("create condition handler must not be called with empty expression")
	}
}

func TestStepHandler_CreateCondition_InvalidConfig(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{err: domainstep.ErrInvalidStepTypeConfig}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateConditionStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_CreateCondition_WorkflowNotFound(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{err: errors.New("workflow not found")}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateConditionStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_CreateCondition_QuotaExceeded(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{err: quotaErr()}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateConditionStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestStepHandler_CreateCondition_InternalError(t *testing.T) {
	createCondition := &mockCreateConditionStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{createCondition: createCondition})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateConditionStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
