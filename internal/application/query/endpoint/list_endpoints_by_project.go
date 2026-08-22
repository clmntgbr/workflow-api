package endpoint

import (
	"context"
	"errors"
	"strings"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type ListEndpointsByProjectQuery struct {
	ProjectID uuid.UUID
	Query          paginate.PaginateQuery
	Methods        []string
}

type ListEndpointsByProjectHandler struct {
	readRepo domainendpoint.EndpointReadRepository
}

func NewListEndpointsByProjectHandler(
	readRepo domainendpoint.EndpointReadRepository,
) *ListEndpointsByProjectHandler {
	return &ListEndpointsByProjectHandler{readRepo: readRepo}
}

func (h *ListEndpointsByProjectHandler) Handle(
	ctx context.Context,
	q ListEndpointsByProjectQuery,
) ([]domainendpoint.EndpointView, int64, error) {
	if q.ProjectID == uuid.Nil {
		return nil, 0, errors.New("projectId is required")
	}

	methods, err := parseMethods(q.Methods)
	if err != nil {
		return nil, 0, err
	}

	views, total, err := h.readRepo.FindByProjectID(ctx, q.ProjectID, domainendpoint.ListEndpointsFilter{
		PaginateQuery: q.Query,
		Methods:       methods,
	})
	if err != nil {
		return nil, 0, errors.New("failed to list endpoints")
	}
	return views, total, nil
}

func parseMethods(raw []string) ([]domainendpoint.Method, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	seen := map[domainendpoint.Method]struct{}{}
	out := make([]domainendpoint.Method, 0, len(raw))
	for _, value := range raw {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			method, err := domainendpoint.ParseMethod(part)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[method]; ok {
				continue
			}
			seen[method] = struct{}{}
			out = append(out, method)
		}
	}
	return out, nil
}
