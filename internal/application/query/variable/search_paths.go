package variable

import (
	"context"

	"go-api/internal/application/query/responsepath"
	"go-api/internal/domain/paginate"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type SearchVariablePathsQuery struct {
	WorkflowID uuid.UUID
	StepID     uuid.UUID
	Query      paginate.PaginateQuery
}

type SearchVariablePathsHandler struct {
	stepRunRepo domainsteprun.StepRunReadRepository
}

func NewSearchVariablePathsHandler(
	stepRunRepo domainsteprun.StepRunReadRepository,
) *SearchVariablePathsHandler {
	return &SearchVariablePathsHandler{stepRunRepo: stepRunRepo}
}

func (h *SearchVariablePathsHandler) Handle(
	ctx context.Context,
	q SearchVariablePathsQuery,
) ([]string, int, error) {
	return responsepath.Search(ctx, h.stepRunRepo, q.WorkflowID, q.StepID, q.Query)
}
