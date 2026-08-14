package read

import (
	"context"
	"errors"
	"time"

	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type variableRow struct {
	ID          uuid.UUID
	Name        string
	Key         string
	Description string
	Path        string
	StepID      uuid.UUID
	WorkflowID  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (variableRow) TableName() string { return "variables" }

type variableReadRepository struct {
	db *gorm.DB
}

func NewVariableReadRepository(db *gorm.DB) domainvariable.VariableReadRepository {
	return &variableReadRepository{db: db}
}

func (r *variableReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainvariable.VariableView, error) {
	var row variableRow
	err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	view := toVariableView(row)
	return &view, nil
}

func (r *variableReadRepository) FindByWorkflowID(
	ctx context.Context,
	workflowID uuid.UUID,
) ([]domainvariable.VariableView, error) {
	var rows []variableRow
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toVariableViews(rows), nil
}

func (r *variableReadRepository) FindByStepID(
	ctx context.Context,
	stepID uuid.UUID,
) ([]domainvariable.VariableView, error) {
	var rows []variableRow
	err := r.db.WithContext(ctx).
		Where("step_id = ?", stepID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toVariableViews(rows), nil
}

func (r *variableReadRepository) FindByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]domainvariable.VariableView, error) {
	if len(ids) == 0 {
		return []domainvariable.VariableView{}, nil
	}
	var rows []variableRow
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toVariableViews(rows), nil
}

func toVariableViews(rows []variableRow) []domainvariable.VariableView {
	out := make([]domainvariable.VariableView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toVariableView(row))
	}
	return out
}

func toVariableView(row variableRow) domainvariable.VariableView {
	return domainvariable.VariableView{
		ID:         row.ID,
		Name:       row.Name,
		Key:        row.Key,
		Description: row.Description,
		Path:       row.Path,
		StepID:     row.StepID,
		WorkflowID: row.WorkflowID,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
