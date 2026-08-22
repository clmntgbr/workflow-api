package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type GetProjectByIDQuery struct {
	ID uuid.UUID
}

type GetProjectByIDHandler struct {
	readRepo domainproject.ProjectReadRepository
}

func NewGetProjectByIDHandler(
	readRepo domainproject.ProjectReadRepository,
) *GetProjectByIDHandler {
	return &GetProjectByIDHandler{readRepo: readRepo}
}

func (h *GetProjectByIDHandler) Handle(
	ctx context.Context,
	q GetProjectByIDQuery,
) (*domainproject.ProjectView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get project")
	}
	if view == nil {
		return nil, errors.New("project not found")
	}
	return view, nil
}
