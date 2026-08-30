package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	domainuser "go-api/internal/domain/user"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

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
