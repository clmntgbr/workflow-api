package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	domainuser "go-api/internal/domain/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// WithLocal injects an arbitrary value into Fiber locals (webhook payloads, etc.).
func WithLocal(key string, value any) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals(key, value)
		return c.Next()
	}
}

func NewTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: validation.FiberErrorHandler,
	})
}

// WithActiveProject injects a user with the given active project into the request context.
func WithActiveProject(userID, projectID uuid.UUID) fiber.Handler {
	return func(c fiber.Ctx) error {
		projectIDCopy := projectID
		httpctx.SetUser(c, domainuser.User{
			ID:              userID,
			ActiveProjectID: &projectIDCopy,
		})
		return c.Next()
	}
}

// WithUserWithoutProject injects a user without an active project.
func WithUserWithoutProject(userID uuid.UUID) fiber.Handler {
	return func(c fiber.Ctx) error {
		httpctx.SetUser(c, domainuser.User{ID: userID})
		return c.Next()
	}
}

func JSONRequest(method, path string, body any) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func DecodeJSON(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func DecodeJSONMap(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	var out map[string]any
	DecodeJSON(t, resp, &out)
	return out
}

func IntPtr(v int) *int       { return &v }
func BoolPtr(v bool) *bool    { return &v }
func StringPtr(v string) *string { return &v }

// MultipartImportRequest builds a multipart request for OpenAPI import endpoints.
func MultipartImportRequest(path string, spec []byte, payload map[string]any) (*http.Request, error) {
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
