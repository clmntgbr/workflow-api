package headertest

import (
	"context"

	queryheader "go-api/internal/application/query/header"
	domainheader "go-api/internal/domain/header"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
)

const (
	suggestPath       = "/headers/suggest"
	suggestValuesPath = "/headers/suggest-values"
)

type mockSuggestHandler struct {
	called      bool
	query       queryheader.SuggestHeadersQuery
	suggestions []domainheader.HeaderSuggestion
	total       int64
	err         error
}

func (m *mockSuggestHandler) Handle(
	_ context.Context,
	q queryheader.SuggestHeadersQuery,
) ([]domainheader.HeaderSuggestion, int64, error) {
	m.called = true
	m.query = q
	return m.suggestions, m.total, m.err
}

type mockSuggestValuesHandler struct {
	called      bool
	query       queryheader.SuggestHeaderValuesQuery
	suggestions []domainheader.HeaderValueSuggestion
	total       int64
	err         error
}

func (m *mockSuggestValuesHandler) Handle(
	_ context.Context,
	q queryheader.SuggestHeaderValuesQuery,
) ([]domainheader.HeaderValueSuggestion, int64, error) {
	m.called = true
	m.query = q
	return m.suggestions, m.total, m.err
}

func newHeaderHandler(
	suggest *mockSuggestHandler,
	suggestValues *mockSuggestValuesHandler,
) *handler.HeaderHandler {
	if suggest == nil {
		suggest = &mockSuggestHandler{}
	}
	if suggestValues == nil {
		suggestValues = &mockSuggestValuesHandler{}
	}
	return handler.NewHeaderHandler(suggest, suggestValues)
}

// mountRoutes registers both header routes behind the given context middleware.
func mountRoutes(h *handler.HeaderHandler, ctx fiber.Handler) *fiber.App {
	app := testutil.NewTestApp()
	if ctx != nil {
		app.Get(suggestPath, ctx, h.Suggest)
		app.Get(suggestValuesPath, ctx, h.SuggestValues)
		return app
	}
	app.Get(suggestPath, h.Suggest)
	app.Get(suggestValuesPath, h.SuggestValues)
	return app
}

func withActiveProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func sampleKeySuggestions() []domainheader.HeaderSuggestion {
	return []domainheader.HeaderSuggestion{
		{Key: "Authorization", Count: 12},
		{Key: "Content-Type", Count: 7},
	}
}

func sampleValueSuggestions() []domainheader.HeaderValueSuggestion {
	return []domainheader.HeaderValueSuggestion{
		{Key: "Authorization", Value: "Bearer {{token-id}}", Count: 5},
		{Key: "Content-Type", Value: "application/json", Count: 3},
	}
}
