package write

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type connectionWriteRepository struct {
	db *gorm.DB
}

func NewConnectionWriteRepository(db *gorm.DB) domainconnection.ConnectionWriteRepository {
	return &connectionWriteRepository{db: db}
}

func (r *connectionWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *connectionWriteRepository) Save(ctx context.Context, conn *domainconnection.Connection) error {
	model := connectionModelFromDomain(conn)
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *connectionWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainconnection.Connection, error) {
	var model ConnectionModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return connectionDomainFromModel(&model), nil
}

func (r *connectionWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&ConnectionModel{}, "id = ?", id).Error
}
