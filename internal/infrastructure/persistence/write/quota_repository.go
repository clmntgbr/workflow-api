package write

import (
	"context"
	"errors"

	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type quotaWriteRepository struct {
	db *gorm.DB
}

func NewQuotaWriteRepository(db *gorm.DB) domainquota.QuotaWriteRepository {
	return &quotaWriteRepository{db: db}
}

func (r *quotaWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *quotaWriteRepository) Save(ctx context.Context, quota *domainquota.Quota) error {
	return DBWithContext(ctx, r.db).Create(quotaModelFromDomain(quota)).Error
}

func (r *quotaWriteRepository) Update(ctx context.Context, quota *domainquota.Quota) error {
	return DBWithContext(ctx, r.db).Save(quotaModelFromDomain(quota)).Error
}

func (r *quotaWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainquota.Quota, error) {
	var model QuotaModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return quotaDomainFromModel(&model), nil
}
