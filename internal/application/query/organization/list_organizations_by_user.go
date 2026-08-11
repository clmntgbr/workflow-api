package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
)

type ListOrganizationsByUserQuery struct {
	UserID uuid.UUID
}

type ListOrganizationsByUserHandler struct {
	readRepo domainorganization.OrganizationReadRepository
}

func NewListOrganizationsByUserHandler(
	readRepo domainorganization.OrganizationReadRepository,
) *ListOrganizationsByUserHandler {
	return &ListOrganizationsByUserHandler{readRepo: readRepo}
}

func (h *ListOrganizationsByUserHandler) Handle(
	ctx context.Context,
	q ListOrganizationsByUserQuery,
) ([]domainorganization.OrganizationView, error) {
	if q.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}
	views, err := h.readRepo.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to list organizations")
	}
	return views, nil
}
