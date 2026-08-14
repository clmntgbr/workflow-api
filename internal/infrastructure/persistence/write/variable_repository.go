package write

import (
	"context"
	"errors"

	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type variableWriteRepository struct {
	db *gorm.DB
}

func NewVariableWriteRepository(db *gorm.DB) domainvariable.VariableWriteRepository {
	return &variableWriteRepository{db: db}
}

func (r *variableWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *variableWriteRepository) Save(ctx context.Context, variable *domainvariable.Variable) error {
	err := DBWithContext(ctx, r.db).Create(variableModelFromDomain(variable)).Error
	if isUniqueViolation(err) {
		return domainvariable.ErrDuplicateKey
	}
	return err
}

func (r *variableWriteRepository) Update(ctx context.Context, variable *domainvariable.Variable) error {
	err := DBWithContext(ctx, r.db).Save(variableModelFromDomain(variable)).Error
	if isUniqueViolation(err) {
		return domainvariable.ErrDuplicateKey
	}
	return err
}

func (r *variableWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&VariableModel{}, "id = ?", id).Error
}

func (r *variableWriteRepository) DeleteByStepID(ctx context.Context, stepID uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&VariableModel{}, "step_id = ?", stepID).Error
}

func (r *variableWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainvariable.Variable, error) {
	var model VariableModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return variableDomainFromModel(&model), nil
}

func isUniqueViolation(err error) bool {
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
