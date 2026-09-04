package assertion

import (
	"context"

	"go-api/internal/application/query/responsepath"
	"go-api/internal/domain/paginate"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type SearchAssertionPathsQuery struct {
	WorkflowID uuid.UUID
	StepID     uuid.UUID
	Query      paginate.PaginateQuery
}

type SearchAssertionPathsHandler struct {
	stepRunRepo domainsteprun.StepRunReadRepository
}

func NewSearchAssertionPathsHandler(
	stepRunRepo domainsteprun.StepRunReadRepository,
) *SearchAssertionPathsHandler {
	return &SearchAssertionPathsHandler{stepRunRepo: stepRunRepo}
}

func (h *SearchAssertionPathsHandler) Handle(
	ctx context.Context,
	q SearchAssertionPathsQuery,
) ([]string, int, error) {
	return responsepath.Search(ctx, h.stepRunRepo, q.WorkflowID, q.StepID, q.Query)
}
