package project

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProjectWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, org *Project) error
	Update(ctx context.Context, org *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	ReplaceMembers(ctx context.Context, projectID uuid.UUID, memberIDs []uuid.UUID) error
}

type ProjectReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*ProjectView, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]ProjectView, error)
}

type ProjectView struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	MemberIDs []uuid.UUID
}
