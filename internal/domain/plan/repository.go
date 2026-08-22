package plan

import (
	"context"
	"time"

	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
)

type PlanWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, plan *Plan) error
	Update(ctx context.Context, plan *Plan) error
	GetByID(ctx context.Context, id uuid.UUID) (*Plan, error)
}

type PlanReadRepository interface {
	FindActive(ctx context.Context) ([]PlanView, error)
	FindAll(ctx context.Context) ([]PlanView, error)
}

type PlanView struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Slug            string
	StripePriceID   string
	IsActive        bool
	BillingInterval BillingInterval
	Price           float64
	Currency        Currency
	QuotaID         uuid.UUID
	Quota           *domainquota.QuotaView
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
