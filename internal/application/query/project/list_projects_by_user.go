package project

import (
	"context"
	"errors"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ListProjectsByUserQuery struct {
	UserID uuid.UUID
}

type ListProjectsByUserHandler struct {
	readRepo domainproject.ProjectReadRepository
}

func NewListProjectsByUserHandler(
	readRepo domainproject.ProjectReadRepository,
) *ListProjectsByUserHandler {
	return &ListProjectsByUserHandler{readRepo: readRepo}
}

func (h *ListProjectsByUserHandler) Handle(
	ctx context.Context,
	q ListProjectsByUserQuery,
) ([]domainproject.ProjectView, error) {
	if q.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}
	views, err := h.readRepo.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to list projects")
	}
	return views, nil
}
