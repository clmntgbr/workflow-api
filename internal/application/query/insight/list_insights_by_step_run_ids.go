package insight

import (
	"context"
	"errors"

	domaininsight "go-api/internal/domain/insight"

	"github.com/google/uuid"
)

type ListInsightsByStepRunIDsQuery struct {
	StepRunIDs []uuid.UUID
}

type ListInsightsByStepRunIDsHandler struct {
	readRepo domaininsight.InsightReadRepository
}

func NewListInsightsByStepRunIDsHandler(
	readRepo domaininsight.InsightReadRepository,
) *ListInsightsByStepRunIDsHandler {
	return &ListInsightsByStepRunIDsHandler{readRepo: readRepo}
}

func (h *ListInsightsByStepRunIDsHandler) Handle(
	ctx context.Context,
	q ListInsightsByStepRunIDsQuery,
) ([]domaininsight.InsightView, error) {
	views, err := h.readRepo.FindByStepRunIDs(ctx, q.StepRunIDs)
	if err != nil {
		return nil, errors.New("failed to list insights")
	}
	return views, nil
}
