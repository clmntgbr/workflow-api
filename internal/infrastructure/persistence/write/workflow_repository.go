package write

import (
	"context"
	"errors"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowWriteRepository struct {
	db *gorm.DB
}

func NewWorkflowWriteRepository(db *gorm.DB) domainworkflow.WorkflowWriteRepository {
	return &workflowWriteRepository{db: db}
}

func (r *workflowWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *workflowWriteRepository) Save(ctx context.Context, workflow *domainworkflow.Workflow) error {
	return DBWithContext(ctx, r.db).Create(workflowModelFromDomain(workflow)).Error
}

func (r *workflowWriteRepository) Update(ctx context.Context, workflow *domainworkflow.Workflow) error {
	return DBWithContext(ctx, r.db).Save(workflowModelFromDomain(workflow)).Error
}

func (r *workflowWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return DBWithContext(ctx, r.db).Delete(&WorkflowModel{}, id).Error
}

func (r *workflowWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainworkflow.Workflow, error) {
	var model WorkflowModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return workflowDomainFromModel(&model), nil
}
