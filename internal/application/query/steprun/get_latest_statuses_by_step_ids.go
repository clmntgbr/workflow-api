package steprun

import (
	"context"

	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type GetLatestStepRunStatusesByStepIDsQuery struct {
	StepIDs []uuid.UUID
}

type GetLatestStepRunStatusesByStepIDsHandler struct {
	readRepo domainsteprun.StepRunReadRepository
}

func NewGetLatestStepRunStatusesByStepIDsHandler(
	readRepo domainsteprun.StepRunReadRepository,
) *GetLatestStepRunStatusesByStepIDsHandler {
	return &GetLatestStepRunStatusesByStepIDsHandler{readRepo: readRepo}
}

func (h *GetLatestStepRunStatusesByStepIDsHandler) Handle(
	ctx context.Context,
	q GetLatestStepRunStatusesByStepIDsQuery,
) (map[uuid.UUID]domainsteprun.Status, error) {
	return h.readRepo.FindLatestStatusByStepIDs(ctx, q.StepIDs)
}
