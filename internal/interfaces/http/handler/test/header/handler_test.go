package headertest

import (
	"errors"
	"net/http"
	"testing"

	"go-api/internal/domain/paginate"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
)

func doGet(t *testing.T, app *fiber.App, path string) *http.Response {
	t.Helper()
	req, err := testutil.JSONRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return resp
}

func TestHeaderHandler_Suggest_Success(t *testing.T) {
	suggest := &mockSuggestHandler{suggestions: sampleKeySuggestions(), total: 2}
	app := mountRoutes(newHeaderHandler(suggest, nil), withActiveProject())

	resp := doGet(t, app, suggestPath+"?page=2&limit=10&search=auth")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	if !suggest.called {
		t.Fatal("expected suggest handler to be called")
	}
	if suggest.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s want %s", suggest.query.ProjectID, testutil.TestProjectID)
	}
	if suggest.query.Search != "auth" {
		t.Fatalf("search: got %q want %q", suggest.query.Search, "auth")
	}
	if suggest.query.Paginate.Page != 2 || suggest.query.Paginate.Limit != 10 {
		t.Fatalf("paginate: got page=%d limit=%d want page=2 limit=10", suggest.query.Paginate.Page, suggest.query.Paginate.Limit)
	}

	var out struct {
		Members    []presenter.HeaderSuggestion `json:"members"`
		Total      int                          `json:"total"`
		Page       int                          `json:"page"`
		Limit      int                          `json:"limit"`
		TotalPages int                          `json:"totalPages"`
	}
	testutil.DecodeJSON(t, resp, &out)

	if len(out.Members) != 2 {
		t.Fatalf("members length: got %d want 2", len(out.Members))
	}
	if out.Members[0].Key != "Authorization" || out.Members[0].Count != 12 {
		t.Fatalf("first member: got %+v", out.Members[0])
	}
	if out.Total != 2 || out.Page != 2 || out.Limit != 10 || out.TotalPages != 1 {
		t.Fatalf("envelope: got total=%d page=%d limit=%d totalPages=%d", out.Total, out.Page, out.Limit, out.TotalPages)
	}
}

// The handler must normalize pagination before calling the query handler,
// otherwise an unbounded limit reaches the repository.
func TestHeaderHandler_Suggest_NormalizesPaginationBeforeHandler(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
	}{
		{name: "defaults", query: "", wantPage: 1, wantLimit: paginate.DefaultLimit},
		{name: "limit above maximum", query: "?limit=5000", wantPage: 1, wantLimit: paginate.DefaultLimit},
		{name: "negative page", query: "?page=-3", wantPage: 1, wantLimit: paginate.DefaultLimit},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			suggest := &mockSuggestHandler{}
			app := mountRoutes(newHeaderHandler(suggest, nil), withActiveProject())

			resp := doGet(t, app, suggestPath+tc.query)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
			}
			if suggest.query.Paginate.Page != tc.wantPage {
				t.Fatalf("page: got %d want %d", suggest.query.Paginate.Page, tc.wantPage)
			}
			if suggest.query.Paginate.Limit != tc.wantLimit {
				t.Fatalf("limit: got %d want %d", suggest.query.Paginate.Limit, tc.wantLimit)
			}
		})
	}
}

func TestHeaderHandler_Suggest_MissingActiveProject(t *testing.T) {
	suggest := &mockSuggestHandler{}
	app := mountRoutes(newHeaderHandler(suggest, nil), testutil.WithUserWithoutProject(testutil.TestUserID))

	resp := doGet(t, app, suggestPath)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggest.called {
		t.Fatal("suggest handler must not be called without active project")
	}
}

func TestHeaderHandler_Suggest_Unauthenticated(t *testing.T) {
	suggest := &mockSuggestHandler{}
	app := mountRoutes(newHeaderHandler(suggest, nil), nil)

	resp := doGet(t, app, suggestPath)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggest.called {
		t.Fatal("suggest handler must not be called without user")
	}
}

func TestHeaderHandler_Suggest_InvalidQuery(t *testing.T) {
	suggest := &mockSuggestHandler{}
	app := mountRoutes(newHeaderHandler(suggest, nil), withActiveProject())

	resp := doGet(t, app, suggestPath+"?page=not-a-number")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggest.called {
		t.Fatal("suggest handler must not be called with invalid query")
	}
}

func TestHeaderHandler_Suggest_HandlerError_Internal(t *testing.T) {
	suggest := &mockSuggestHandler{err: errors.New("database unavailable")}
	app := mountRoutes(newHeaderHandler(suggest, nil), withActiveProject())

	resp := doGet(t, app, suggestPath)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to fetch header suggestions" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestHeaderHandler_SuggestValues_Success(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{suggestions: sampleValueSuggestions(), total: 2}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), withActiveProject())

	resp := doGet(t, app, suggestValuesPath+"?key=Authorization&search=bearer&page=1&limit=25")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}

	if !suggestValues.called {
		t.Fatal("expected suggest values handler to be called")
	}
	if suggestValues.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s want %s", suggestValues.query.ProjectID, testutil.TestProjectID)
	}
	if suggestValues.query.Key != "Authorization" {
		t.Fatalf("key: got %q want %q", suggestValues.query.Key, "Authorization")
	}
	if suggestValues.query.Search != "bearer" {
		t.Fatalf("search: got %q want %q", suggestValues.query.Search, "bearer")
	}
	if suggestValues.query.Paginate.Limit != 25 {
		t.Fatalf("limit: got %d want 25", suggestValues.query.Paginate.Limit)
	}

	var out struct {
		Members []presenter.HeaderValueSuggestion `json:"members"`
		Total   int                               `json:"total"`
	}
	testutil.DecodeJSON(t, resp, &out)

	if len(out.Members) != 2 {
		t.Fatalf("members length: got %d want 2", len(out.Members))
	}
	// The raw template must be returned, not an interpolated value.
	if out.Members[0].Value != "Bearer {{token-id}}" {
		t.Fatalf("value: got %q", out.Members[0].Value)
	}
	if out.Members[0].Key != "Authorization" {
		t.Fatalf("key: got %q", out.Members[0].Key)
	}
	if out.Total != 2 {
		t.Fatalf("total: got %d want 2", out.Total)
	}
}

// key is optional: when absent the query handler searches across every key.
func TestHeaderHandler_SuggestValues_WithoutKey(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{suggestions: sampleValueSuggestions(), total: 2}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), withActiveProject())

	resp := doGet(t, app, suggestValuesPath+"?search=json")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if suggestValues.query.Key != "" {
		t.Fatalf("key: got %q want empty", suggestValues.query.Key)
	}
	if suggestValues.query.Search != "json" {
		t.Fatalf("search: got %q want %q", suggestValues.query.Search, "json")
	}
}

func TestHeaderHandler_SuggestValues_NormalizesPaginationBeforeHandler(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), withActiveProject())

	resp := doGet(t, app, suggestValuesPath+"?limit=5000&page=0")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if suggestValues.query.Paginate.Limit != paginate.DefaultLimit {
		t.Fatalf("limit: got %d want %d", suggestValues.query.Paginate.Limit, paginate.DefaultLimit)
	}
	if suggestValues.query.Paginate.Page != 1 {
		t.Fatalf("page: got %d want 1", suggestValues.query.Paginate.Page)
	}
}

func TestHeaderHandler_SuggestValues_MissingActiveProject(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), testutil.WithUserWithoutProject(testutil.TestUserID))

	resp := doGet(t, app, suggestValuesPath+"?key=Authorization")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggestValues.called {
		t.Fatal("handler must not be called without active project")
	}
}

func TestHeaderHandler_SuggestValues_Unauthenticated(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), nil)

	resp := doGet(t, app, suggestValuesPath)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggestValues.called {
		t.Fatal("handler must not be called without user")
	}
}

func TestHeaderHandler_SuggestValues_InvalidQuery(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), withActiveProject())

	resp := doGet(t, app, suggestValuesPath+"?limit=not-a-number")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if suggestValues.called {
		t.Fatal("handler must not be called with invalid query")
	}
}

func TestHeaderHandler_SuggestValues_HandlerError_Internal(t *testing.T) {
	suggestValues := &mockSuggestValuesHandler{err: errors.New("database unavailable")}
	app := mountRoutes(newHeaderHandler(nil, suggestValues), withActiveProject())

	resp := doGet(t, app, suggestValuesPath+"?key=Authorization")
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to fetch header value suggestions" {
		t.Fatalf("message: got %v", body["message"])
	}
}
