package handler

import (
	"context"

	queryheader "go-api/internal/application/query/header"
	domainheader "go-api/internal/domain/header"
)

type headerSuggestHandler interface {
	Handle(ctx context.Context, q queryheader.SuggestHeadersQuery) ([]domainheader.HeaderSuggestion, int64, error)
}

type headerValueSuggestHandler interface {
	Handle(ctx context.Context, q queryheader.SuggestHeaderValuesQuery) ([]domainheader.HeaderValueSuggestion, int64, error)
}
