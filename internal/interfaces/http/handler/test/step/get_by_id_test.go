package steptest

import (
	"errors"
	"net/http"
	"testing"

	domainstep "go-api/internal/domain/step"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

func TestStepHandler_GetByID_Success(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{sampleHTTPStepView()},
		errs:  []error{nil},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	var out presenter.StepDetailResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testutil.TestStepID.String() {
		t.Fatalf("response id: got %s", out.ID)
	}
}

func TestStepHandler_GetByID_NoActiveProject(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), testutil.WithUserWithoutProject(testutil.TestUserID), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_GetByID_InvalidWorkflowID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(
		t,
		http.MethodGet,
		"/workflows/bad/steps/"+testutil.TestStepID.String(),
		nil,
	))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_GetByID_InvalidStepID(t *testing.T) {
	h := newStepHandler(stepMocks{})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepsBasePath()+"/bad", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestStepHandler_GetByID_NotFound(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		errs: []error{errors.New("step not found")},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_GetByID_WrongProject(t *testing.T) {
	view := sampleHTTPStepView()
	view.ProjectID = otherProjectID
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{view},
		errs:  []error{nil},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_GetByID_WrongWorkflow(t *testing.T) {
	view := sampleHTTPStepView()
	view.WorkflowID = uuid.New()
	getByID := &mockGetStepByIDHandler{
		views: []*domainstep.StepView{view},
		errs:  []error{nil},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestStepHandler_GetByID_InternalError(t *testing.T) {
	getByID := &mockGetStepByIDHandler{
		errs: []error{errors.New("database unavailable")},
	}
	h := newStepHandler(stepMocks{getByID: getByID})

	app := testutil.NewTestApp()
	app.Get(stepItemRoute(), activeProject(), h.GetByID)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, stepItemPath(testutil.TestStepID), nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
