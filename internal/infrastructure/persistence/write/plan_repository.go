package write

import (
	"context"
	"errors"

	domainplan "go-api/internal/domain/plan"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type planWriteRepository struct {
	db *gorm.DB
}

func NewPlanWriteRepository(db *gorm.DB) domainplan.PlanWriteRepository {
	return &planWriteRepository{db: db}
}

func (r *planWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *planWriteRepository) Save(ctx context.Context, plan *domainplan.Plan) error {
	err := DBWithContext(ctx, r.db).Create(planModelFromDomain(plan)).Error
	if isPlanUniqueViolation(err) {
		return domainplan.ErrDuplicateSlug
	}
	return err
}

func (r *planWriteRepository) Update(ctx context.Context, plan *domainplan.Plan) error {
	err := DBWithContext(ctx, r.db).Save(planModelFromDomain(plan)).Error
	if isPlanUniqueViolation(err) {
		return domainplan.ErrDuplicateSlug
	}
	return err
}

func (r *planWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func (r *planWriteRepository) GetBySlug(ctx context.Context, slug string) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).First(&model, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func (r *planWriteRepository) GetByStripePriceID(ctx context.Context, stripePriceID string) (*domainplan.Plan, error) {
	var model PlanModel
	err := DBWithContext(ctx, r.db).Where("stripe_price_id = ?", stripePriceID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return planDomainFromModel(&model), nil
}

func isPlanUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
