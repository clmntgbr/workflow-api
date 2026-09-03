package steptest

import (
	"errors"
	"net/http"
	"testing"

	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

func TestStepHandler_UpdatePosition_Success(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{result: sampleHTTPStepEntity()}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !updatePosition.called {
		t.Fatal("expected update position handler to be called")
	}
	if updatePosition.cmd.Position.X != 50 {
		t.Fatalf("position x: got %f want 50", updatePosition.cmd.Position.X)
	}

	var out presenter.StepDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestStepID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestStepHandler_UpdatePosition_Unauthorized(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if updatePosition.called {
		t.Fatal("update position handler must not be called without user")
	}
}

func TestStepHandler_UpdatePosition_NoActiveProject(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdatePosition_InvalidWorkflowID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		"/workflows/bad/steps/"+testutil.TestStepID.String()+"/position",
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdatePosition_InvalidStepID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepsBasePath()+"/bad/position",
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_UpdatePosition_InvalidBody(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		"not-json",
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if updatePosition.called {
		t.Fatal("update position handler must not be called on validation error")
	}
}

// (0,0) must stay valid: the position tags are bounds, not `required`.
func TestStepHandler_UpdatePosition_ZeroOriginIsValid(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{result: sampleHTTPStepEntity()}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		map[string]any{"position": map[string]any{"x": 0.0, "y": 0.0}},
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !updatePosition.called {
		t.Fatal("expected update position handler to be called for origin position")
	}
}

func TestStepHandler_UpdatePosition_OutOfBounds(t *testing.T) {
	tests := []struct {
		name      string
		position  map[string]any
		wantField string
	}{
		{name: "x too small", position: map[string]any{"x": -2000000.0, "y": 0.0}, wantField: "x"},
		{name: "x too large", position: map[string]any{"x": 2000000.0, "y": 0.0}, wantField: "x"},
		{name: "y too large", position: map[string]any{"x": 0.0, "y": 2000000.0}, wantField: "y"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updatePosition := &mockUpdateStepPositionHandler{}
			h := newStepHandler(stepMocks{updatePosition: updatePosition})

			app := testutil.NewTestApp()
			app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

			resp, err := app.Test(mustJSONRequest(
				t,
				http.MethodPut,
				stepPositionPath(testutil.TestStepID),
				map[string]any{"position": tc.position},
			))
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
			}
			if updatePosition.called {
				t.Fatal("update position handler must not be called on validation error")
			}

			body := testutil.DecodeJSONMap(t, resp)
			fieldErrors, ok := body["errors"].(map[string]any)
			if !ok {
				t.Fatalf("errors: got %T want map", body["errors"])
			}
			if _, found := fieldErrors[tc.wantField]; !found {
				t.Fatalf("errors: missing field %q, got %v", tc.wantField, fieldErrors)
			}
		})
	}
}

func TestStepHandler_UpdatePosition_NotFound(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{err: errors.New("step not found")}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_UpdatePosition_InternalError(t *testing.T) {
	updatePosition := &mockUpdateStepPositionHandler{err: errors.New("database unavailable")}
	h := newStepHandler(stepMocks{updatePosition: updatePosition})

	app := testutil.NewTestApp()
	app.Put(stepPositionRoute(), activeProject(), h.UpdatePosition)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodPut,
		stepPositionPath(testutil.TestStepID),
		validUpdatePositionBody(),
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
