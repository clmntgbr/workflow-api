package assertion

import (
	"context"
	"errors"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
)

type GetAssertionByIDQuery struct {
	ID         uuid.UUID
	WorkflowID uuid.UUID
}

type GetAssertionByIDHandler struct {
	repo domainassertion.AssertionReadRepository
}

func NewGetAssertionByIDHandler(repo domainassertion.AssertionReadRepository) *GetAssertionByIDHandler {
	return &GetAssertionByIDHandler{repo: repo}
}

func (h *GetAssertionByIDHandler) Handle(
	ctx context.Context,
	q GetAssertionByIDQuery,
) (*domainassertion.AssertionView, error) {
	view, err := h.repo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	if view == nil || view.WorkflowID != q.WorkflowID {
		return nil, domainassertion.ErrNotFound
	}
	return view, nil
}

type ListAssertionsByStepQuery struct {
	StepID uuid.UUID
}

type ListAssertionsByStepHandler struct {
	repo domainassertion.AssertionReadRepository
}

func NewListAssertionsByStepHandler(repo domainassertion.AssertionReadRepository) *ListAssertionsByStepHandler {
	return &ListAssertionsByStepHandler{repo: repo}
}

func (h *ListAssertionsByStepHandler) Handle(
	ctx context.Context,
	q ListAssertionsByStepQuery,
) ([]domainassertion.AssertionView, error) {
	if q.StepID == uuid.Nil {
		return nil, errors.New("stepId is required")
	}
	return h.repo.FindByStepID(ctx, q.StepID)
}
