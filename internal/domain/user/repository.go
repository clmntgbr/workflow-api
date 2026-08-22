package user

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type UserWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByClerkID(ctx context.Context, clerkID string) error
	GetByClerkID(ctx context.Context, clerkID string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (*User, error)
}

type UserReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*UserView, error)
	FindByClerkID(ctx context.Context, clerkID string) (*UserView, error)
}

type UserView struct {
	ID                   uuid.UUID
	ClerkID              string
	FirstName            string
	LastName             string
	Email                string
	Banned               bool
	ActiveProjectID *uuid.UUID
	SubscriptionID       *uuid.UUID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
