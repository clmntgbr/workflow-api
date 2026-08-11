package organization

import (
	"context"
	"errors"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
)

type GetOrganizationByIDQuery struct {
	ID uuid.UUID
}

type GetOrganizationByIDHandler struct {
	readRepo domainorganization.OrganizationReadRepository
}

func NewGetOrganizationByIDHandler(
	readRepo domainorganization.OrganizationReadRepository,
) *GetOrganizationByIDHandler {
	return &GetOrganizationByIDHandler{readRepo: readRepo}
}

func (h *GetOrganizationByIDHandler) Handle(
	ctx context.Context,
	q GetOrganizationByIDQuery,
) (*domainorganization.OrganizationView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get organization")
	}
	if view == nil {
		return nil, errors.New("organization not found")
	}
	return view, nil
}
