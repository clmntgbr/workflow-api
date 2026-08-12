package step

import (
	"context"
	"errors"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type GetStepByIDQuery struct {
	ID uuid.UUID
}

type GetStepByIDHandler struct {
	readRepo domainstep.StepReadRepository
}

func NewGetStepByIDHandler(readRepo domainstep.StepReadRepository) *GetStepByIDHandler {
	return &GetStepByIDHandler{readRepo: readRepo}
}

func (h *GetStepByIDHandler) Handle(
	ctx context.Context,
	q GetStepByIDQuery,
) (*domainstep.StepView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if view == nil || view.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	return view, nil
}
