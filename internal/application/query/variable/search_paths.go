package variable

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"

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
