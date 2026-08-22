package endpoint

import (
	"context"
	"time"

	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type EndpointWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, endpoint *Endpoint) error
	Update(ctx context.Context, endpoint *Endpoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)
}

type ListEndpointsFilter struct {
	paginate.PaginateQuery
	Methods []Method
}

type EndpointReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*EndpointView, error)
	FindByProjectID(
		ctx context.Context,
		projectID uuid.UUID,
		filter ListEndpointsFilter,
	) ([]EndpointView, int64, error)
}

type EndpointView struct {
	ID             uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         Method
	Headers        map[string]string
	Query          httpquery.Params
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         Status
	ProjectID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
