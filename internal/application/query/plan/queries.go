package plan

import (
	"context"
	"errors"

	domainplan "go-api/internal/domain/plan"
)

type ListActivePlansQuery struct{}

type ListActivePlansHandler struct {
	readRepo domainplan.PlanReadRepository
}

func NewListActivePlansHandler(readRepo domainplan.PlanReadRepository) *ListActivePlansHandler {
	return &ListActivePlansHandler{readRepo: readRepo}
}

func (h *ListActivePlansHandler) Handle(ctx context.Context, _ ListActivePlansQuery) ([]domainplan.PlanView, error) {
	views, err := h.readRepo.FindActive(ctx)
	if err != nil {
		return nil, errors.New("failed to list plans")
	}
	return views, nil
}
