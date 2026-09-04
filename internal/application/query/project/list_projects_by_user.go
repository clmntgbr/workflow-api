package project

import (
	"context"
	"errors"

	"go-api/internal/domain/paginate"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ListProjectsByUserQuery struct {
	UserID uuid.UUID
	Query  paginate.PaginateQuery
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
) ([]domainproject.ProjectView, int64, error) {
	if q.UserID == uuid.Nil {
		return nil, 0, errors.New("userId is required")
	}
	views, total, err := h.readRepo.FindPageByUserID(ctx, q.UserID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list projects")
	}
	return views, total, nil
}
