package plan

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID          uuid.UUID
	Name        string
	Description string
	Slug        string

	StripePriceID string

	IsActive bool

	BillingInterval BillingInterval
	Price           float64
	Currency        Currency

	QuotaID uuid.UUID

	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewPlanParams struct {
	Name            string
	Description     string
	Slug            string
	StripePriceID   string
	IsActive        bool
	BillingInterval BillingInterval
	Price           float64
	Currency        Currency
	QuotaID         uuid.UUID
}

func NewPlan(p NewPlanParams) (*Plan, error) {
	if err := validatePlanInput(p.Name, p.Slug, p.StripePriceID, p.BillingInterval, p.Currency, p.QuotaID); err != nil {
		return nil, err
	}
	interval := p.BillingInterval
	if interval == "" {
		interval = BillingIntervalMonth
	}
	currency := p.Currency
	if currency == "" {
		currency = CurrencyEUR
	}

	now := time.Now().UTC()
	return &Plan{
		ID:              uuid.New(),
		Name:            strings.TrimSpace(p.Name),
		Description:     strings.TrimSpace(p.Description),
		Slug:            strings.TrimSpace(p.Slug),
		StripePriceID:   strings.TrimSpace(p.StripePriceID),
		IsActive:        p.IsActive,
		BillingInterval: interval,
		Price:           p.Price,
		Currency:        currency,
		QuotaID:         p.QuotaID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

type UpdatePlanParams struct {
	Name            string
	Description     string
	Slug            string
	StripePriceID   string
	IsActive        bool
	BillingInterval BillingInterval
	Price           float64
	Currency        Currency
	QuotaID         uuid.UUID
}

func (p *Plan) ApplyUpdate(params UpdatePlanParams) error {
	if err := validatePlanInput(
		params.Name,
		params.Slug,
		params.StripePriceID,
		params.BillingInterval,
		params.Currency,
		params.QuotaID,
	); err != nil {
		return err
	}
	interval := params.BillingInterval
	if interval == "" {
		interval = BillingIntervalMonth
	}
	currency := params.Currency
	if currency == "" {
		currency = CurrencyEUR
	}

	p.Name = strings.TrimSpace(params.Name)
	p.Description = strings.TrimSpace(params.Description)
	p.Slug = strings.TrimSpace(params.Slug)
	p.StripePriceID = strings.TrimSpace(params.StripePriceID)
	p.IsActive = params.IsActive
	p.BillingInterval = interval
	p.Price = params.Price
	p.Currency = currency
	p.QuotaID = params.QuotaID
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (p *Plan) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now().UTC()
}

func (p *Plan) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now().UTC()
}
