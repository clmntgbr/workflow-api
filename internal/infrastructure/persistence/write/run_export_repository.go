package write

import (
	"context"
	"errors"

	domainrunexport "go-api/internal/domain/runexport"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type runExportWriteRepository struct {
	db *gorm.DB
}

func NewRunExportWriteRepository(db *gorm.DB) domainrunexport.WriteRepository {
	return &runExportWriteRepository{db: db}
}

func (r *runExportWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *runExportWriteRepository) Save(ctx context.Context, job *domainrunexport.RunExport) error {
	return DBWithContext(ctx, r.db).Create(runExportModelFromDomain(job)).Error
}

func (r *runExportWriteRepository) Update(ctx context.Context, job *domainrunexport.RunExport) error {
	return DBWithContext(ctx, r.db).Save(runExportModelFromDomain(job)).Error
}

func (r *runExportWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainrunexport.RunExport, error) {
	var model RunExportModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return runExportDomainFromModel(&model), nil
}
