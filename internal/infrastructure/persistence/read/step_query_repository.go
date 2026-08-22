package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stepRow struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	ProjectID uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        dbtype.JSONB
	QueryParams    dbtype.JSONB
	Body           dbtype.JSONB
	Timeout        int  `gorm:"column:timeout_ms"`
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int  `gorm:"column:retry_delay_ms"`
	StepIndex      string
	ExecutionOrder int
	TreeIndex      int
	PositionX      float64
	PositionY      float64
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (stepRow) TableName() string { return "steps" }

var stepSelectColumns = []string{
	"id", "workflow_id", "endpoint_id", "project_id",
	"name", "description", "url", "method", "headers", "query_params", "body",
	"timeout_ms", "retry_on_failure", "retry_count", "retry_delay_ms",
	"step_index", "execution_order", "tree_index", "position_x", "position_y",
	"status", "created_at", "updated_at",
}

type stepReadRepository struct {
	db *gorm.DB
}

func NewStepReadRepository(db *gorm.DB) domainstep.StepReadRepository {
	return &stepReadRepository{db: db}
}

func (r *stepReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainstep.StepView, error) {
	var row stepRow
	err := r.db.WithContext(ctx).
		Select(stepSelectColumns).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toStepView(row)
}

func (r *stepReadRepository) FindByWorkflowID(
	ctx context.Context,
	workflowID uuid.UUID,
) ([]domainstep.StepView, error) {
	var rows []stepRow
	err := r.db.WithContext(ctx).
		Model(&stepRow{}).
		Select(stepSelectColumns).
		Where("workflow_id = ? AND status <> ?", workflowID, domainstep.StatusDeleted).
		Order("execution_order ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domainstep.StepView, 0, len(rows))
	for _, row := range rows {
		view, err := toStepView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func toStepView(row stepRow) (*domainstep.StepView, error) {
	headers := map[string]string{}
	if len(row.Headers) > 0 {
		if err := json.Unmarshal(row.Headers, &headers); err != nil {
			return nil, err
		}
	}

	query := httpquery.Empty()
	if len(row.QueryParams) > 0 {
		if err := json.Unmarshal(row.QueryParams, &query); err != nil {
			return nil, err
		}
	}

	body := map[string]any{}
	if len(row.Body) > 0 {
		if err := json.Unmarshal(row.Body, &body); err != nil {
			return nil, err
		}
	}

	return &domainstep.StepView{
		ID:             row.ID,
		WorkflowID:     row.WorkflowID,
		EndpointID:     row.EndpointID,
		ProjectID: row.ProjectID,
		Name:           row.Name,
		Description:    row.Description,
		URL:            row.URL,
		Method:         row.Method,
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        row.Timeout,
		RetryOnFailure: row.RetryOnFailure,
		RetryCount:     row.RetryCount,
		RetryDelay:     row.RetryDelay,
		Index:          row.StepIndex,
		ExecutionOrder: row.ExecutionOrder,
		TreeIndex:      row.TreeIndex,
		Position: domainstep.Position{
			X: row.PositionX,
			Y: row.PositionY,
		},
		Status:    domainstep.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}
