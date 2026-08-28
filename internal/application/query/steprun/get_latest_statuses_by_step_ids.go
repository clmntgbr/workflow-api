package steprun

import (
	"context"

	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type GetLatestStepRunStatusesByStepIDsQuery struct {
	WorkflowID uuid.UUID
	StepIDs    []uuid.UUID
}

type GetLatestStepRunStatusesByStepIDsHandler struct {
	stepRunReadRepo     domainsteprun.StepRunReadRepository
	workflowRunReadRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewGetLatestStepRunStatusesByStepIDsHandler(
	stepRunReadRepo domainsteprun.StepRunReadRepository,
	workflowRunReadRepo domainworkflowrun.WorkflowRunReadRepository,
) *GetLatestStepRunStatusesByStepIDsHandler {
	return &GetLatestStepRunStatusesByStepIDsHandler{
		stepRunReadRepo:     stepRunReadRepo,
		workflowRunReadRepo: workflowRunReadRepo,
	}
}

func (h *GetLatestStepRunStatusesByStepIDsHandler) Handle(
	ctx context.Context,
	q GetLatestStepRunStatusesByStepIDsQuery,
) (map[uuid.UUID]domainsteprun.Status, error) {
	if q.WorkflowID != uuid.Nil {
		activeRun, err := h.workflowRunReadRepo.FindInProgressByWorkflowID(ctx, q.WorkflowID)
		if err != nil {
			return nil, err
		}
		if activeRun != nil {
			return h.stepRunReadRepo.FindStatusByWorkflowRunIDAndStepIDs(ctx, activeRun.ID, q.StepIDs)
		}
	}

	return h.stepRunReadRepo.FindLatestStatusByStepIDs(ctx, q.StepIDs)
}
