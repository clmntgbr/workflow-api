package assertion

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"

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
	if q.StepID == uuid.Nil {
		return nil, 0, errors.New("stepId is required")
	}

	q.Query.Normalize()

	stepRun, err := h.stepRunRepo.FindLatestCompletedByStepID(ctx, q.StepID)
	if err != nil {
		return nil, 0, errors.New("failed to get latest step run")
	}
	if stepRun == nil || stepRun.ResponseSnapshot == nil || stepRun.ResponseSnapshot.Body == nil {
		return []string{}, 0, nil
	}
	if stepRun.WorkflowID != q.WorkflowID {
		return nil, 0, errors.New("step run not found")
	}

	allPaths := domainvariable.ExtractResponsePaths(stepRun.ResponseSnapshot.Body, q.Query.Search)
	total := len(allPaths)

	start := q.Query.Offset()
	if start >= total {
		return []string{}, total, nil
	}
	end := start + q.Query.Limit
	if end > total {
		end = total
	}

	return allPaths[start:end], total, nil
}
