package subscription

import (
	"context"
	"time"

	domainplan "go-api/internal/domain/plan"

	"github.com/google/uuid"
)

type SubscriptionWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, subscription *Subscription) error
	Update(ctx context.Context, subscription *Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*Subscription, error)
	GetByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*Subscription, error)
}

type SubscriptionReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*SubscriptionView, error)
	FindByStripeSubscriptionID(ctx context.Context, stripeSubscriptionID string) (*SubscriptionView, error)
	FindByStripeCustomerID(ctx context.Context, stripeCustomerID string) (*SubscriptionView, error)
}

type SubscriptionView struct {
	ID                   uuid.UUID
	PlanID               uuid.UUID
	Plan                 *domainplan.PlanView
	StripeCustomerID     string
	StripeSubscriptionID string
	Status               Status
	StartDate            time.Time
	EndDate              time.Time
	CancelAtPeriodEnd    bool
	QuotaPeriodStart     time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
