package write

import (
	"context"
	"errors"

	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowRunWriteRepository struct {
	db *gorm.DB
}

func NewWorkflowRunWriteRepository(db *gorm.DB) domainworkflowrun.WorkflowRunWriteRepository {
	return &workflowRunWriteRepository{db: db}
}

func (r *workflowRunWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *workflowRunWriteRepository) Save(ctx context.Context, run *domainworkflowrun.WorkflowRun) error {
	model, err := workflowRunModelFromDomain(run)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *workflowRunWriteRepository) Update(ctx context.Context, run *domainworkflowrun.WorkflowRun) error {
	model, err := workflowRunModelFromDomain(run)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *workflowRunWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainworkflowrun.WorkflowRun, error) {
	var model WorkflowRunModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return workflowRunDomainFromModel(&model)
}

func (r *workflowRunWriteRepository) HasInProgress(ctx context.Context, workflowID uuid.UUID) (bool, error) {
	var count int64
	err := DBWithContext(ctx, r.db).
		Model(&WorkflowRunModel{}).
		Where("workflow_id = ? AND status IN ?", workflowID, []string{
			string(domainworkflowrun.StatusPending),
			string(domainworkflowrun.StatusRunning),
		}).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
