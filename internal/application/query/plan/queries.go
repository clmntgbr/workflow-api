package plan

import (
	"context"
	"errors"

	domainplan "go-api/internal/domain/plan"

	"github.com/google/uuid"
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

type GetPlanByIDQuery struct {
	ID uuid.UUID
}

type GetPlanByIDHandler struct {
	readRepo domainplan.PlanReadRepository
}

func NewGetPlanByIDHandler(readRepo domainplan.PlanReadRepository) *GetPlanByIDHandler {
	return &GetPlanByIDHandler{readRepo: readRepo}
}

func (h *GetPlanByIDHandler) Handle(ctx context.Context, q GetPlanByIDQuery) (*domainplan.PlanView, error) {
	if q.ID == uuid.Nil {
		return nil, errors.New("plan id is required")
	}
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get plan")
	}
	if view == nil {
		return nil, domainplan.ErrNotFound
	}
	return view, nil
}
