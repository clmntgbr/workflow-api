package header

import (
	"context"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type HeaderSuggestion struct {
	Key   string
	Count int64
}

type HeaderValueSuggestion struct {
	Key   string
	Value string
	Count int64
}

type HeaderSuggestionFilter struct {
	ProjectID uuid.UUID
	Search    string
	Paginate  paginate.PaginateQuery
}

type HeaderValueSuggestionFilter struct {
	ProjectID uuid.UUID
	Key       string // Optional - if empty, search all keys
	Search    string
	Paginate  paginate.PaginateQuery
}

type ReadRepository interface {
	FindSuggestions(
		ctx context.Context,
		filter HeaderSuggestionFilter,
	) ([]HeaderSuggestion, int64, error)
	FindValuesByKey(
		ctx context.Context,
		filter HeaderValueSuggestionFilter,
	) ([]HeaderValueSuggestion, int64, error)
}
