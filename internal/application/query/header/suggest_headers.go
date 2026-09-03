package header

import (
	"context"
	"errors"

	domainheader "go-api/internal/domain/header"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type SuggestHeadersQuery struct {
	ProjectID uuid.UUID
	Search    string
	Paginate  paginate.PaginateQuery
}

type SuggestHeadersHandler struct {
	readRepo domainheader.ReadRepository
}

func NewSuggestHeadersHandler(
	readRepo domainheader.ReadRepository,
) *SuggestHeadersHandler {
	return &SuggestHeadersHandler{readRepo: readRepo}
}

func (h *SuggestHeadersHandler) Handle(
	ctx context.Context,
	q SuggestHeadersQuery,
) ([]domainheader.HeaderSuggestion, int64, error) {
	if q.ProjectID == uuid.Nil {
		return nil, 0, errors.New("projectId is required")
	}

	suggestions, total, err := h.readRepo.FindSuggestions(ctx, domainheader.HeaderSuggestionFilter{
		ProjectID: q.ProjectID,
		Search:    q.Search,
		Paginate:  q.Paginate,
	})
	if err != nil {
		return nil, 0, errors.New("failed to fetch header suggestions")
	}

	return suggestions, total, nil
}
