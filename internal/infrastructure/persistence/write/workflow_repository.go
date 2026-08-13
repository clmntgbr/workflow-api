package write

import (
	"context"
	"errors"
	"time"

	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *workflowWriteRepository) ClaimDueForExecution(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]*domainworkflow.Workflow, error) {
	if limit <= 0 {
		limit = 100
	}

	var claimed []*domainworkflow.Workflow
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []WorkflowModel
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", domainworkflow.StatusActive).
			Where("next_run_at IS NOT NULL AND next_run_at <= ?", now).
			Where(
				"EXISTS (SELECT 1 FROM steps WHERE steps.workflow_id = workflows.id AND steps.status <> ?)",
				domainstep.StatusDeleted,
			).
			Order("next_run_at ASC").
			Limit(limit).
			Find(&models).Error
		if err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}

		claimed = make([]*domainworkflow.Workflow, 0, len(models))
		for i := range models {
			workflow := workflowDomainFromModel(&models[i])
			workflow.AdvanceAfterScheduledStart(now)
			if err := tx.Save(workflowModelFromDomain(workflow)).Error; err != nil {
				return err
			}
			claimed = append(claimed, workflow)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}
