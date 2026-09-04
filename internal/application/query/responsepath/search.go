// Package responsepath resolves the JSON paths available in a step's latest
// response, used to populate variable and assertion path pickers.
package responsepath

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

// Search returns one page of response paths extracted from the latest completed
// run of the given step, along with the total before paging.
//
// Paging is applied in memory: the paths are derived from a single response
// snapshot, so there is no query to push the offset down into.
func Search(
	ctx context.Context,
	stepRunRepo domainsteprun.StepRunReadRepository,
	workflowID uuid.UUID,
	stepID uuid.UUID,
	query paginate.PaginateQuery,
) ([]string, int, error) {
	if stepID == uuid.Nil {
		return nil, 0, errors.New("stepId is required")
	}

	query.Normalize()

	stepRun, err := stepRunRepo.FindLatestCompletedByStepID(ctx, stepID)
	if err != nil {
		return nil, 0, errors.New("failed to get latest step run")
	}
	if stepRun == nil || stepRun.ResponseSnapshot == nil || stepRun.ResponseSnapshot.Body == nil {
		return []string{}, 0, nil
	}
	if stepRun.WorkflowID != workflowID {
		return nil, 0, errors.New("step run not found")
	}

	allPaths := domainvariable.ExtractResponsePaths(stepRun.ResponseSnapshot.Body, query.Search)
	total := len(allPaths)

	start := query.Offset()
	if start >= total {
		return []string{}, total, nil
	}
	end := start + query.Limit
	if end > total {
		end = total
	}

	return allPaths[start:end], total, nil
}
