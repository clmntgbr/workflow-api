package read

import (
	"context"
	"errors"
	"time"

	domainplan "go-api/internal/domain/plan"
	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type planRow struct {
	ID              uuid.UUID
	Name            string
	Description     string
	Slug            string
	StripePriceID   string
	IsActive        bool
	BillingInterval string
	Price           float64
	Currency        string
	QuotaID         uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (planRow) TableName() string { return "plans" }

type planReadRepository struct {
	db        *gorm.DB
	quotaRead domainquota.QuotaReadRepository
}

func NewPlanReadRepository(db *gorm.DB, quotaRead domainquota.QuotaReadRepository) domainplan.PlanReadRepository {
	return &planReadRepository{db: db, quotaRead: quotaRead}
}

func (r *planReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainplan.PlanView, error) {
	var row planRow
	err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toPlanViewWithQuota(ctx, row)
}

func (r *planReadRepository) FindActive(ctx context.Context) ([]domainplan.PlanView, error) {
	var rows []planRow
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return r.toPlanViewsWithQuota(ctx, rows)
}

func (r *planReadRepository) FindAll(ctx context.Context) ([]domainplan.PlanView, error) {
	var rows []planRow
	err := r.db.WithContext(ctx).
		Order("price ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return r.toPlanViewsWithQuota(ctx, rows)
}

func (r *planReadRepository) toPlanViewsWithQuota(ctx context.Context, rows []planRow) ([]domainplan.PlanView, error) {
	if len(rows) == 0 {
		return []domainplan.PlanView{}, nil
	}
	quotaIDs := make([]uuid.UUID, 0, len(rows))
	seen := map[uuid.UUID]struct{}{}
	for _, row := range rows {
		if _, ok := seen[row.QuotaID]; ok {
			continue
		}
		seen[row.QuotaID] = struct{}{}
		quotaIDs = append(quotaIDs, row.QuotaID)
	}
	quotas, err := r.quotaRead.FindByIDs(ctx, quotaIDs)
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]domainquota.QuotaView, len(quotas))
	for _, quota := range quotas {
		byID[quota.ID] = quota
	}

	out := make([]domainplan.PlanView, 0, len(rows))
	for _, row := range rows {
		view := toPlanView(row)
		if quota, ok := byID[row.QuotaID]; ok {
			q := quota
			view.Quota = &q
		}
		out = append(out, view)
	}
	return out, nil
}

func (r *planReadRepository) toPlanViewWithQuota(ctx context.Context, row planRow) (*domainplan.PlanView, error) {
	view := toPlanView(row)
	quota, err := r.quotaRead.FindByID(ctx, row.QuotaID)
	if err != nil {
		return nil, err
	}
	view.Quota = quota
	return &view, nil
}

func toPlanView(row planRow) domainplan.PlanView {
	return domainplan.PlanView{
		ID:              row.ID,
		Name:            row.Name,
		Description:     row.Description,
		Slug:            row.Slug,
		StripePriceID:   row.StripePriceID,
		IsActive:        row.IsActive,
		BillingInterval: domainplan.BillingInterval(row.BillingInterval),
		Price:           row.Price,
		Currency:        domainplan.Currency(row.Currency),
		QuotaID:         row.QuotaID,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}
