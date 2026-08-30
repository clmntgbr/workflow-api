package endpointtest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"testing"

	endpointcmd "go-api/internal/application/command/endpoint"
	domainendpoint "go-api/internal/domain/endpoint"
	domainuser "go-api/internal/domain/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type importHandlerStub struct{}

func (importHandlerStub) Handle(
	_ context.Context,
	_ endpointcmd.ImportEndpointsFromOpenAPICommand,
) ([]domainendpoint.Endpoint, error) {
	return nil, errors.New("unexpected import call")
}

func TestEndpointHandler_ImportFromOpenAPI_ReadSpecErrors(t *testing.T) {
	tests := []struct {
		name        string
		readErr     error
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "file too large",
			readErr:     handler.ErrOpenAPIFileTooLargeForTest(),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "OpenAPI file is too large",
		},
		{
			name:        "read failed",
			readErr:     handler.ErrOpenAPIFileReadForTest(),
			wantStatus:  http.StatusBadRequest,
			wantMessage: "Failed to read OpenAPI file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			restore := handler.SetReadOpenAPISpecFromMultipartFileForTest(
				func(*multipart.FileHeader, int64) ([]byte, error) {
					return nil, tc.readErr
				},
			)
			t.Cleanup(restore)

			h := handler.NewEndpointHandler(
				nil,
				importHandlerStub{},
				nil,
				nil,
				nil,
				nil,
			)

			app := fiber.New(fiber.Config{ErrorHandler: validation.FiberErrorHandler})
			userID := uuid.New()
			projectID := uuid.New()
			app.Post("/endpoints/import", func(c fiber.Ctx) error {
				projectIDCopy := projectID
				httpctx.SetUser(c, domainuser.User{
					ID:              userID,
					ActiveProjectID: &projectIDCopy,
				})
				return c.Next()
			}, h.ImportFromOpenAPI)

			req, err := multipartImportRequest(
				"/endpoints/import",
				[]byte(`{"openapi":"3.0.3"}`),
				map[string]any{
					"baseUrl":        "https://api.example.com",
					"status":         "active",
					"timeout":        30000,
					"retryOnFailure": false,
					"retryCount":     0,
					"retryDelay":     10000,
				},
			)
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status: got %d want %d", resp.StatusCode, tc.wantStatus)
			}

			var body map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			_ = resp.Body.Close()
			if body["message"] != tc.wantMessage {
				t.Fatalf("message: got %q want %q", body["message"], tc.wantMessage)
			}
		})
	}
}

func multipartImportRequest(path string, spec []byte, payload map[string]any) (*http.Request, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filePart, err := writer.CreateFormFile("file", "openapi.json")
	if err != nil {
		return nil, err
	}
	if _, err := filePart.Write(spec); err != nil {
		return nil, err
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	if err := writer.WriteField("payload", string(payloadJSON)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, path, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}
