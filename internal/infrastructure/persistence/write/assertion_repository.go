package write

import (
	"context"
	"errors"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type assertionWriteRepository struct {
	db *gorm.DB
}

func NewAssertionWriteRepository(db *gorm.DB) domainassertion.AssertionWriteRepository {
	return &assertionWriteRepository{db: db}
}

func (r *assertionWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *assertionWriteRepository) Save(ctx context.Context, assertion *domainassertion.Assertion) error {
	return DBWithContext(ctx, r.db).Create(assertionModelFromDomain(assertion)).Error
}

func (r *assertionWriteRepository) Update(ctx context.Context, assertion *domainassertion.Assertion) error {
	return DBWithContext(ctx, r.db).Save(assertionModelFromDomain(assertion)).Error
}

func (r *assertionWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&AssertionModel{}, "id = ?", id).Error
}

func (r *assertionWriteRepository) DeleteByStepID(ctx context.Context, stepID uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&AssertionModel{}, "step_id = ?", stepID).Error
}

func (r *assertionWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainassertion.Assertion, error) {
	var model AssertionModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return assertionDomainFromModel(&model), nil
}
