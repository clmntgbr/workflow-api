package write

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type endpointWriteRepository struct {
	db *gorm.DB
}

func NewEndpointWriteRepository(db *gorm.DB) domainendpoint.EndpointWriteRepository {
	return &endpointWriteRepository{db: db}
}

func (r *endpointWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *endpointWriteRepository) Save(ctx context.Context, endpoint *domainendpoint.Endpoint) error {
	model, err := endpointModelFromDomain(endpoint)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *endpointWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainendpoint.Endpoint, error) {
	var model EndpointModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return endpointDomainFromModel(&model)
}
