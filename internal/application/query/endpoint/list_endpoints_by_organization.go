package endpoint

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ListEndpointsByOrganizationQuery struct {
	OrganizationID uuid.UUID
	Query          paginate.PaginateQuery
}

type ListEndpointsByOrganizationHandler struct {
	readRepo domainendpoint.EndpointReadRepository
}

func NewListEndpointsByOrganizationHandler(
	readRepo domainendpoint.EndpointReadRepository,
) *ListEndpointsByOrganizationHandler {
	return &ListEndpointsByOrganizationHandler{readRepo: readRepo}
}

func (h *ListEndpointsByOrganizationHandler) Handle(
	ctx context.Context,
	q ListEndpointsByOrganizationQuery,
) ([]domainendpoint.EndpointView, int64, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, 0, errors.New("organizationId is required")
	}
	views, total, err := h.readRepo.FindByOrganizationID(ctx, q.OrganizationID, q.Query)
	if err != nil {
		return nil, 0, errors.New("failed to list endpoints")
	}
	return views, total, nil
}
