package workflowtest

import (
	"errors"
	"net/http"
	"testing"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/workflowio"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/interfaces/http/testutil"
)

func TestWorkflowHandler_Export_Success(t *testing.T) {
	exportedAt := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	exportH := &mockExportWorkflowHandler{
		result: &workflowio.Document{
			FormatVersion: 1,
			ExportedAt:    exportedAt,
			Workflow: workflowio.Workflow{
				Name:                 "Order Flow",
				ScheduleType:         "none",
				ScheduleTimezone:     "UTC",
				Concurrency:          1,
				NotificationsEnabled: true,
			},
			Steps:       []workflowio.Step{},
			Connections: []workflowio.Connection{},
			Variables:   []workflowio.Variable{},
			Assertions:  []workflowio.Assertion{},
		},
	}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", activeProject(), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !exportH.called {
		t.Fatal("expected export handler to be called")
	}
	if exportH.query.ID != testutil.TestWorkflowID {
		t.Fatalf("workflow id: got %s", exportH.query.ID)
	}
	if exportH.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", exportH.query.ProjectID)
	}

	var out workflowio.Document
	testutil.DecodeJSON(t, resp, &out)
	if out.FormatVersion != 1 || out.Workflow.Name != "Order Flow" {
		t.Fatalf("unexpected export payload: %+v", out)
	}
}

func TestWorkflowHandler_Export_Unauthorized(t *testing.T) {
	exportH := &mockExportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if exportH.called {
		t.Fatal("export handler must not be called without user")
	}
}

func TestWorkflowHandler_Export_MissingActiveProject(t *testing.T) {
	exportH := &mockExportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", testutil.WithUserWithoutProject(testutil.TestUserID), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Export_InvalidID(t *testing.T) {
	exportH := &mockExportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", activeProject(), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/bad-id/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if exportH.called {
		t.Fatal("export handler must not be called with invalid id")
	}
}

func TestWorkflowHandler_Export_NotFound(t *testing.T) {
	exportH := &mockExportWorkflowHandler{err: errors.New("workflow not found")}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", activeProject(), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWorkflowHandler_Export_HandlerError(t *testing.T) {
	exportH := &mockExportWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", activeProject(), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Export_InvalidDocument(t *testing.T) {
	exportH := &mockExportWorkflowHandler{err: workflowio.Invalid("connection source step is missing from export")}
	h := newWorkflowHandlerWithIO(exportH, nil)

	app := testutil.NewTestApp()
	app.Get("/workflows/:workflowId/export", activeProject(), h.Export)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/workflows/"+testutil.TestWorkflowID.String()+"/export", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Import_Success(t *testing.T) {
	importH := &mockImportWorkflowHandler{result: sampleWorkflowEntity()}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !importH.called {
		t.Fatal("expected import handler to be called")
	}
	if importH.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", importH.cmd.UserID)
	}
	if importH.cmd.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", importH.cmd.ProjectID)
	}
	if importH.cmd.Document.FormatVersion != 1 {
		t.Fatalf("format version: got %d", importH.cmd.Document.FormatVersion)
	}
	if importH.cmd.Document.Workflow.Name != "Order sync" {
		t.Fatalf("workflow name: got %s", importH.cmd.Document.Workflow.Name)
	}
}

func TestWorkflowHandler_Import_Unauthorized(t *testing.T) {
	importH := &mockImportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if importH.called {
		t.Fatal("import handler must not be called without user")
	}
}

func TestWorkflowHandler_Import_MissingActiveProject(t *testing.T) {
	importH := &mockImportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", testutil.WithUserWithoutProject(testutil.TestUserID), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Import_InvalidBody(t *testing.T) {
	importH := &mockImportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", map[string]any{
		"formatVersion": 1,
		"workflow":      map[string]any{},
	}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called with invalid body")
	}
}

func TestWorkflowHandler_Import_UnsupportedFormat(t *testing.T) {
	importH := &mockImportWorkflowHandler{}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	body := validImportWorkflowBody()
	body["formatVersion"] = 2

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", body))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called with unsupported format")
	}
}

func TestWorkflowHandler_Import_QuotaExceeded(t *testing.T) {
	importH := &mockImportWorkflowHandler{err: cmdquota.ErrWorkflowQuotaExceeded}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestWorkflowHandler_Import_NotAllowed(t *testing.T) {
	importH := &mockImportWorkflowHandler{err: cmdquota.ErrWorkflowImportNotAllowed}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}

func TestWorkflowHandler_Import_InvalidDocument(t *testing.T) {
	importH := &mockImportWorkflowHandler{err: workflowio.Invalid("unknown connection sourceRef")}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestWorkflowHandler_Import_HandlerError(t *testing.T) {
	importH := &mockImportWorkflowHandler{err: errors.New("database unavailable")}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestWorkflowHandler_Import_InvalidSchedule(t *testing.T) {
	importH := &mockImportWorkflowHandler{err: domainworkflow.ErrInvalidScheduleTimezone}
	h := newWorkflowHandlerWithIO(nil, importH)

	app := testutil.NewTestApp()
	app.Post("/workflows/import", activeProject(), h.Import)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/workflows/import", validImportWorkflowBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func validImportWorkflowBody() map[string]any {
	return map[string]any{
		"formatVersion": 1,
		"workflow": map[string]any{
			"name":                 "Order sync",
			"scheduleType":         "none",
			"scheduleTimezone":     "UTC",
			"concurrency":          1,
			"notificationsEnabled": true,
		},
		"steps": []map[string]any{
			{
				"ref":            "step-1",
				"type":           "http",
				"name":           "Login",
				"url":            "https://api.example.com/login",
				"method":         "POST",
				"timeout":        30000,
				"retryOnFailure": true,
				"retryCount":     0,
				"retryDelay":     10000,
				"position":       map[string]any{"x": 100, "y": 200},
			},
		},
	}
}
