package steptest

import (
	"errors"
	"net/http"
	"testing"

	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_Delete_Success(t *testing.T) {
	deleteH := &mockDeleteStepHandler{}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteH.called {
		t.Fatal("expected delete handler to be called")
	}
	if deleteH.cmd.ID != testutil.TestStepID {
		t.Fatalf("step id: got %s", deleteH.cmd.ID)
	}
}

func TestStepHandler_Delete_Unauthorized(t *testing.T) {
	deleteH := &mockDeleteStepHandler{}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called without user")
	}
}

func TestStepHandler_Delete_NoActiveProject(t *testing.T) {
	deleteH := &mockDeleteStepHandler{}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called without active project")
	}
}

func TestStepHandler_Delete_InvalidWorkflowID(t *testing.T) {
	deleteH := &mockDeleteStepHandler{}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodDelete,
		"/workflows/bad/steps/"+testutil.TestStepID.String(),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called with invalid workflow id")
	}
}

func TestStepHandler_Delete_InvalidStepID(t *testing.T) {
	deleteH := &mockDeleteStepHandler{}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepsBasePath()+"/bad", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if deleteH.called {
		t.Fatal("delete handler must not be called with invalid step id")
	}
}

func TestStepHandler_Delete_NotFound(t *testing.T) {
	deleteH := &mockDeleteStepHandler{err: errors.New("step not found")}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_Delete_InternalError(t *testing.T) {
	deleteH := &mockDeleteStepHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{deleteH: deleteH})

	app := testutil.NewTestApp()
	app.Delete(stepItemRoute(), activeProject(), h.Delete)

	resp, err := app.Test(mustJSONRequest(t, http.MethodDelete, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
