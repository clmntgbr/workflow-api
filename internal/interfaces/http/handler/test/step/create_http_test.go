package steptest

import (
	"errors"
	"net/http"
	"testing"

	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_CreateHTTP_Success(t *testing.T) {
	create := &mockCreateStepHandler{result: sampleHTTPStepEntity()}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
	if create.cmd.EndpointID != testutil.TestEndpointID {
		t.Fatalf("endpoint id: got %s", create.cmd.EndpointID)
	}

	var out presenter.StepDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestStepID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestStepHandler_CreateHTTP_DefaultType(t *testing.T) {
	create := &mockCreateStepHandler{result: sampleHTTPStepEntity()}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := validCreateHTTPStepBody()
	delete(body, "type")

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
}

func TestStepHandler_CreateHTTP_Unauthorized(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
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

func TestStepHandler_CreateHTTP_NoActiveProject(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called without active project")
	}
}

func TestStepHandler_CreateHTTP_InvalidWorkflowID(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/not-a-uuid/steps", validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid workflow id")
	}
}

func TestStepHandler_CreateHTTP_InvalidBody(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), "not-json"))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called on invalid body")
	}
}

func TestStepHandler_CreateHTTP_MissingEndpointID(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), map[string]any{
		"position": map[string]any{"x": 0.0, "y": 0.0},
	}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called without endpoint id")
	}
}

func TestStepHandler_CreateHTTP_InvalidEndpointID(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := validCreateHTTPStepBody()
	body["endpointId"] = "bad-id"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid endpoint id")
	}
}

func TestStepHandler_CreateHTTP_WorkflowNotFound(t *testing.T) {
	create := &mockCreateStepHandler{err: errors.New("workflow not found")}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_CreateHTTP_EndpointNotFound(t *testing.T) {
	create := &mockCreateStepHandler{err: errors.New("endpoint not found")}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_CreateHTTP_QuotaExceeded(t *testing.T) {
	create := &mockCreateStepHandler{err: quotaErr()}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestStepHandler_CreateHTTP_InternalError(t *testing.T) {
	create := &mockCreateStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), validCreateHTTPStepBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestStepHandler_CreateHTTP_InvalidTypeValidation(t *testing.T) {
	create := &mockCreateStepHandler{}
	h := newStepHandler(stepMocks{create: create})

	app := testutil.NewTestApp()
	app.Post(stepsRoute(), activeProject(), h.Create)

	body := validCreateHTTPStepBody()
	body["type"] = "invalid"

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, stepsBasePath(), body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called on validation error")
	}
}
