package organization

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type OrganizationWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, org *Organization) error
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	ReplaceMembers(ctx context.Context, organizationID uuid.UUID, memberIDs []uuid.UUID) error
}

type OrganizationReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*OrganizationView, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]OrganizationView, error)
}

type OrganizationView struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	MemberIDs []uuid.UUID
}
