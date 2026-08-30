package endpointtest

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"testing"

	cmdquota "go-api/internal/application/command/quota"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

const minimalOpenAPISpec = `{
  "openapi": "3.0.3",
  "info": {"title": "Test API", "version": "1.0.0"},
  "paths": {
    "/users": {
      "get": {"operationId": "listUsers", "responses": {"200": {"description": "ok"}}}
    }
  }
}`

func validImportPayload() map[string]any {
	return map[string]any{
		"baseUrl":        "https://api.example.com",
		"status":         "active",
		"timeout":        30000,
		"retryOnFailure": false,
		"retryCount":     0,
		"retryDelay":     10000,
	}
}

func TestEndpointHandler_ImportFromOpenAPI_Success(t *testing.T) {
	entity := sampleEndpointEntity()
	importH := &mockImportEndpointHandler{result: []domainendpoint.Endpoint{*entity}}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if !importH.called {
		t.Fatal("expected import handler to be called")
	}
	if importH.cmd.UserID != testUserID {
		t.Fatalf("user id: got %s want %s", importH.cmd.UserID, testUserID)
	}
	if importH.cmd.ProjectID != testProjectID {
		t.Fatalf("project id: got %s want %s", importH.cmd.ProjectID, testProjectID)
	}
	if importH.cmd.BaseURL != "https://api.example.com" {
		t.Fatalf("base url: got %q", importH.cmd.BaseURL)
	}
	if string(importH.cmd.Spec) != minimalOpenAPISpec {
		t.Fatal("expected spec bytes to be forwarded to handler")
	}

	var out []presenter.EndpointListResponse
	testutil.DecodeJSON(t, resp, &out)
	if len(out) != 1 {
		t.Fatalf("response length: got %d want 1", len(out))
	}
}

func TestEndpointHandler_ImportFromOpenAPI_Unauthorized(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
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

func TestEndpointHandler_ImportFromOpenAPI_MissingActiveProject(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithUserWithoutProject(testUserID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called without active project")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_MissingFile(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("payload", `{"baseUrl":"https://api.example.com"}`)
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/endpoints/import", &body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called without file")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_MissingPayload(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	filePart, err := writer.CreateFormFile("file", "openapi.json")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := filePart.Write([]byte(minimalOpenAPISpec)); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	_ = writer.Close()

	req, err := http.NewRequest(http.MethodPost, "/endpoints/import", &body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called without payload")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_InvalidPayloadJSON(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	filePart, _ := writer.CreateFormFile("file", "openapi.json")
	_, _ = filePart.Write([]byte(minimalOpenAPISpec))
	_ = writer.WriteField("payload", "not-json")
	_ = writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/endpoints/import", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called with invalid payload json")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_InvalidPayloadValidation(t *testing.T) {
	importH := &mockImportEndpointHandler{}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	payload := validImportPayload()
	delete(payload, "baseUrl")

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		payload,
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if importH.called {
		t.Fatal("import handler must not be called on validation error")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_NoOperations(t *testing.T) {
	importH := &mockImportEndpointHandler{err: domainendpoint.ErrNoOperations}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_TooManyOperations(t *testing.T) {
	importH := &mockImportEndpointHandler{err: domainendpoint.ErrTooManyOperations}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_InvalidEndpointURL(t *testing.T) {
	importH := &mockImportEndpointHandler{err: domainendpoint.ErrInvalidEndpointURL}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_InvalidOpenAPI(t *testing.T) {
	importH := &mockImportEndpointHandler{err: domainendpoint.ErrInvalidOpenAPI}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if !importH.called {
		t.Fatal("expected import handler to be called")
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_Internal(t *testing.T) {
	importH := &mockImportEndpointHandler{err: errors.New("database unavailable")}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestEndpointHandler_ImportFromOpenAPI_HandlerError_QuotaExceeded(t *testing.T) {
	importH := &mockImportEndpointHandler{err: cmdquota.ErrEndpointQuotaExceeded}
	h := newEndpointHandler(nil, nil, nil, nil, nil, importH)

	app := testutil.NewTestApp()
	app.Post("/endpoints/import", testutil.WithActiveProject(testUserID, testProjectID), h.ImportFromOpenAPI)

	req, err := testutil.MultipartImportRequest(
		"/endpoints/import",
		[]byte(minimalOpenAPISpec),
		validImportPayload(),
	)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusForbidden)
	}
}
