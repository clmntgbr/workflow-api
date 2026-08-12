package endpoint

import (
	"context"
	"time"

	"go-api/internal/domain/paginate"

	"github.com/google/uuid"
)

type EndpointWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, endpoint *Endpoint) error
	GetByID(ctx context.Context, id uuid.UUID) (*Endpoint, error)
}

type EndpointReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*EndpointView, error)
	FindByOrganizationID(
		ctx context.Context,
		organizationID uuid.UUID,
		query paginate.PaginateQuery,
	) ([]EndpointView, int64, error)
}

type EndpointView struct {
	ID             uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         Method
	Headers        map[string]string
	Query          map[string]string
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         Status
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
