package header

import (
	"context"
	"errors"

	domainheader "go-api/internal/domain/header"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type SuggestHeaderValuesQuery struct {
	ProjectID uuid.UUID
	Key       string
	Search    string
	Paginate  paginate.PaginateQuery
}

type SuggestHeaderValuesHandler struct {
	readRepo domainheader.ReadRepository
}

func NewSuggestHeaderValuesHandler(
	readRepo domainheader.ReadRepository,
) *SuggestHeaderValuesHandler {
	return &SuggestHeaderValuesHandler{readRepo: readRepo}
}

func (h *SuggestHeaderValuesHandler) Handle(
	ctx context.Context,
	q SuggestHeaderValuesQuery,
) ([]domainheader.HeaderValueSuggestion, int64, error) {
	if q.ProjectID == uuid.Nil {
		return nil, 0, errors.New("projectId is required")
	}

	suggestions, total, err := h.readRepo.FindValuesByKey(ctx, domainheader.HeaderValueSuggestionFilter{
		ProjectID: q.ProjectID,
		Key:       q.Key, // Optional
		Search:    q.Search,
		Paginate:  q.Paginate,
	})
	if err != nil {
		return nil, 0, errors.New("failed to fetch header value suggestions")
	}

	return suggestions, total, nil
}
