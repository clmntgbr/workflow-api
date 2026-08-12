package endpoint

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"

	"github.com/google/uuid"
)

type GetEndpointByIDQuery struct {
	ID uuid.UUID
}

type GetEndpointByIDHandler struct {
	readRepo domainendpoint.EndpointReadRepository
}

func NewGetEndpointByIDHandler(readRepo domainendpoint.EndpointReadRepository) *GetEndpointByIDHandler {
	return &GetEndpointByIDHandler{readRepo: readRepo}
}

func (h *GetEndpointByIDHandler) Handle(
	ctx context.Context,
	q GetEndpointByIDQuery,
) (*domainendpoint.EndpointView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get endpoint")
	}
	if view == nil || view.Status == domainendpoint.StatusDeleted {
		return nil, errors.New("endpoint not found")
	}
	return view, nil
}
