package read

import (
	"context"
	"errors"
	"time"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type assertionRow struct {
	ID            uuid.UUID
	Name          string
	Description   string
	Source        string
	Path          *string
	Operator      string
	ExpectedValue *string
	StepID        uuid.UUID
	WorkflowID    uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (assertionRow) TableName() string { return "assertions" }

type assertionReadRepository struct {
	db *gorm.DB
}

func NewAssertionReadRepository(db *gorm.DB) domainassertion.AssertionReadRepository {
	return &assertionReadRepository{db: db}
}

func (r *assertionReadRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainassertion.AssertionView, error) {
	var row assertionRow
	err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	view := toAssertionView(row)
	return &view, nil
}

func (r *assertionReadRepository) FindByWorkflowID(
	ctx context.Context,
	workflowID uuid.UUID,
) ([]domainassertion.AssertionView, error) {
	var rows []assertionRow
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toAssertionViews(rows), nil
}

func (r *assertionReadRepository) FindByStepID(
	ctx context.Context,
	stepID uuid.UUID,
) ([]domainassertion.AssertionView, error) {
	var rows []assertionRow
	err := r.db.WithContext(ctx).
		Where("step_id = ?", stepID).
		Order("created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toAssertionViews(rows), nil
}

func toAssertionViews(rows []assertionRow) []domainassertion.AssertionView {
	items := make([]domainassertion.AssertionView, 0, len(rows))
	for _, row := range rows {
		items = append(items, toAssertionView(row))
	}
	return items
}

func toAssertionView(row assertionRow) domainassertion.AssertionView {
	path := ""
	if row.Path != nil {
		path = *row.Path
	}
	expected := ""
	if row.ExpectedValue != nil {
		expected = *row.ExpectedValue
	}
	return domainassertion.AssertionView{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		Source:        domainassertion.AssertionSource(row.Source),
		Path:          path,
		Operator:      domainassertion.AssertionOperator(row.Operator),
		ExpectedValue: expected,
		StepID:        row.StepID,
		WorkflowID:    row.WorkflowID,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
