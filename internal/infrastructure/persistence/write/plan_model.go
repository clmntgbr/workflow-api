package write

import (
	"time"

	domainplan "go-api/internal/domain/plan"

	"github.com/google/uuid"
)

type PlanModel struct {
	ID              uuid.UUID `gorm:"column:id;primaryKey"`
	Name            string    `gorm:"column:name"`
	Description     string    `gorm:"column:description"`
	Slug            string    `gorm:"column:slug"`
	StripePriceID   string    `gorm:"column:stripe_price_id"`
	IsActive        bool      `gorm:"column:is_active"`
	BillingInterval string    `gorm:"column:billing_interval"`
	Price           float64   `gorm:"column:price;type:decimal(10,2)"`
	Currency        string    `gorm:"column:currency"`
	QuotaID         uuid.UUID `gorm:"column:quota_id"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (PlanModel) TableName() string { return "plans" }

func planModelFromDomain(p *domainplan.Plan) *PlanModel {
	return &PlanModel{
		ID:              p.ID,
		Name:            p.Name,
		Description:     p.Description,
		Slug:            p.Slug,
		StripePriceID:   p.StripePriceID,
		IsActive:        p.IsActive,
		BillingInterval: string(p.BillingInterval),
		Price:           p.Price,
		Currency:        string(p.Currency),
		QuotaID:         p.QuotaID,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

func planDomainFromModel(m *PlanModel) *domainplan.Plan {
	return &domainplan.Plan{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		Slug:            m.Slug,
		StripePriceID:   m.StripePriceID,
		IsActive:        m.IsActive,
		BillingInterval: domainplan.BillingInterval(m.BillingInterval),
		Price:           m.Price,
		Currency:        domainplan.Currency(m.Currency),
		QuotaID:         m.QuotaID,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
