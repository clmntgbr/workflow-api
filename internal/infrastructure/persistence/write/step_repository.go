package write

import (
	"context"
	"errors"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stepWriteRepository struct {
	db *gorm.DB
}

func NewStepWriteRepository(db *gorm.DB) domainstep.StepWriteRepository {
	return &stepWriteRepository{db: db}
}

func (r *stepWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *stepWriteRepository) Save(ctx context.Context, step *domainstep.Step) error {
	model, err := stepModelFromDomain(step)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *stepWriteRepository) Update(ctx context.Context, step *domainstep.Step) error {
	model, err := stepModelFromDomain(step)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *stepWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainstep.Step, error) {
	var model StepModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return stepDomainFromModel(&model)
}
