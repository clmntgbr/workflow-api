package quotatest

import (
	"errors"
	"net/http"
	"testing"

	cmdquota "go-api/internal/application/command/quota"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
)

func TestRespondQuotaError_Mapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "subscription not found",
			err:        querysubscription.ErrSubscriptionNotFound,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "active project required",
			err:        querysubscription.ErrActiveProjectRequired,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "endpoint quota exceeded",
			err:        cmdquota.ErrEndpointQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "workflow quota exceeded",
			err:        cmdquota.ErrWorkflowQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "step quota exceeded",
			err:        cmdquota.ErrStepQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "variable quota exceeded",
			err:        cmdquota.ErrVariableQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "assertion quota exceeded",
			err:        cmdquota.ErrAssertionQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "workflow run quota exceeded",
			err:        cmdquota.ErrWorkflowRunQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "concurrent run quota exceeded",
			err:        cmdquota.ErrConcurrentRunQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "project quota exceeded",
			err:        cmdquota.ErrProjectQuotaExceeded,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "nil error",
			err:        nil,
			wantStatus: 0,
		},
		{
			name:       "unmapped error",
			err:        errors.New("something else"),
			wantStatus: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := testutil.NewTestApp()
			app.Get("/quota", func(c fiber.Ctx) error {
				handled, err := handler.RespondQuotaErrorForTest(c, tc.err)
				if err != nil {
					return err
				}
				if handled {
					return nil
				}
				return c.SendStatus(http.StatusTeapot)
			})

			req, err := testutil.JSONRequest(http.MethodGet, "/quota", nil)
			if err != nil {
				t.Fatalf("build request: %v", err)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("perform request: %v", err)
			}

			if tc.wantStatus == 0 {
				if resp.StatusCode != http.StatusTeapot {
					t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusTeapot)
				}
				return
			}
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status: got %d want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}
