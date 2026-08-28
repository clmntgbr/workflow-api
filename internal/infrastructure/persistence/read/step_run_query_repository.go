package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type stepRunRow struct {
	ID                   uuid.UUID
	WorkflowRunID        uuid.UUID
	StepID               uuid.UUID
	WorkflowID           uuid.UUID
	EndpointID           *uuid.UUID
	ProjectID            uuid.UUID
	StepType             string `gorm:"column:step_type"`
	DelayDurationSeconds *int
	Name                 string
	Description        string
	URL                string
	Method             string
	Headers            dbtype.JSONB
	QueryParams        dbtype.JSONB
	Body               dbtype.JSONB
	Timeout            int `gorm:"column:timeout_ms"`
	RetryOnFailure     bool
	RetryCount         int
	RetryDelay         int `gorm:"column:retry_delay_ms"`
	StepIndex          string
	ExecutionOrder     int
	TreeIndex          int
	PositionX          float64
	PositionY          float64
	Status             string
	Attempt            int
	VariableExtracts   dbtype.JSONB `gorm:"column:variable_extracts"`
	Assertions         dbtype.JSONB `gorm:"column:assertions"`
	ResponseSnapshot   dbtype.JSONB
	ExtractedVariables dbtype.JSONB `gorm:"column:extracted_variables"`
	AssertionsResult   dbtype.JSONB `gorm:"column:assertions_result"`
	StartedAt          *time.Time
	FinishedAt         *time.Time
	ResumeAt           *time.Time
	Error              string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (stepRunRow) TableName() string { return "step_runs" }

var stepRunSelectColumns = []string{
	"id", "workflow_run_id", "step_id", "workflow_id", "endpoint_id", "project_id",
	"step_type", "delay_duration_seconds",
	"name", "description", "url", "method", "headers", "query_params", "body",
	"timeout_ms", "retry_on_failure", "retry_count", "retry_delay_ms",
	"step_index", "execution_order", "tree_index", "position_x", "position_y",
	"status", "attempt", "variable_extracts", "assertions", "response_snapshot", "extracted_variables",
	"assertions_result", "started_at", "finished_at", "resume_at", "error", "created_at", "updated_at",
}

type stepRunReadRepository struct {
	db *gorm.DB
}

func NewStepRunReadRepository(db *gorm.DB) domainsteprun.StepRunReadRepository {
	return &stepRunReadRepository{db: db}
}

func (r *stepRunReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainsteprun.StepRunView, error) {
	var row stepRunRow
	err := r.db.WithContext(ctx).
		Select(stepRunSelectColumns).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toStepRunView(row)
}

func (r *stepRunReadRepository) FindByWorkflowRunID(
	ctx context.Context,
	workflowRunID uuid.UUID,
) ([]domainsteprun.StepRunView, error) {
	var rows []stepRunRow
	err := r.db.WithContext(ctx).
		Model(&stepRunRow{}).
		Select(stepRunSelectColumns).
		Where("workflow_run_id = ?", workflowRunID).
		Order("execution_order ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domainsteprun.StepRunView, 0, len(rows))
	for _, row := range rows {
		view, err := toStepRunView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (r *stepRunReadRepository) FindByWorkflowRunIDs(
	ctx context.Context,
	workflowRunIDs []uuid.UUID,
) ([]domainsteprun.StepRunView, error) {
	if len(workflowRunIDs) == 0 {
		return []domainsteprun.StepRunView{}, nil
	}

	var rows []stepRunRow
	err := r.db.WithContext(ctx).
		Model(&stepRunRow{}).
		Select(stepRunSelectColumns).
		Where("workflow_run_id IN ?", workflowRunIDs).
		Order("execution_order ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domainsteprun.StepRunView, 0, len(rows))
	for _, row := range rows {
		view, err := toStepRunView(row)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (r *stepRunReadRepository) FindLatestCompletedByStepID(
	ctx context.Context,
	stepID uuid.UUID,
) (*domainsteprun.StepRunView, error) {
	var row stepRunRow
	err := r.db.WithContext(ctx).
		Select(stepRunSelectColumns).
		Where("step_id = ? AND status = ?", stepID, string(domainsteprun.StatusSuccess)).
		Order("finished_at DESC NULLS LAST, created_at DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toStepRunView(row)
}

func (r *stepRunReadRepository) FindLatestStatusByStepIDs(
	ctx context.Context,
	stepIDs []uuid.UUID,
) (map[uuid.UUID]domainsteprun.Status, error) {
	out := make(map[uuid.UUID]domainsteprun.Status, len(stepIDs))
	if len(stepIDs) == 0 {
		return out, nil
	}

	type latestStatusRow struct {
		StepID uuid.UUID
		Status string
	}

	var rows []latestStatusRow
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT DISTINCT ON (step_id) step_id, status
			FROM step_runs
			WHERE step_id IN ?
			ORDER BY step_id, created_at DESC
		`, stepIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		out[row.StepID] = domainsteprun.Status(row.Status)
	}
	return out, nil
}

func toStepRunView(row stepRunRow) (*domainsteprun.StepRunView, error) {
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

	var response *domainsteprun.ResponseSnapshot
	if len(row.ResponseSnapshot) > 0 {
		var snapshot domainsteprun.ResponseSnapshot
		if err := json.Unmarshal(row.ResponseSnapshot, &snapshot); err != nil {
			return nil, err
		}
		response = &snapshot
	}

	extracts := []domainsteprun.VariableExtract{}
	if len(row.VariableExtracts) > 0 {
		if err := json.Unmarshal(row.VariableExtracts, &extracts); err != nil {
			return nil, err
		}
	}

	assertions := []domainassertion.Snapshot{}
	if len(row.Assertions) > 0 {
		if err := json.Unmarshal(row.Assertions, &assertions); err != nil {
			return nil, err
		}
	}

	extracted := map[string]any{}
	if len(row.ExtractedVariables) > 0 {
		if err := json.Unmarshal(row.ExtractedVariables, &extracted); err != nil {
			return nil, err
		}
	}

	assertionsResult := []domainassertion.Result{}
	if len(row.AssertionsResult) > 0 {
		if err := json.Unmarshal(row.AssertionsResult, &assertionsResult); err != nil {
			return nil, err
		}
	}

	stepType := domainstep.Type(row.StepType)
	if stepType == "" {
		stepType = domainstep.TypeHTTP
	}

	return &domainsteprun.StepRunView{
		ID:                   row.ID,
		WorkflowRunID:        row.WorkflowRunID,
		StepID:               row.StepID,
		WorkflowID:           row.WorkflowID,
		EndpointID:           row.EndpointID,
		ProjectID:            row.ProjectID,
		StepType:             stepType,
		DelayDurationSeconds: intValueOrZero(row.DelayDurationSeconds),
		Name:                 row.Name,
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
		Status:             domainsteprun.Status(row.Status),
		Attempt:            row.Attempt,
		VariableExtracts:   extracts,
		Assertions:         assertions,
		ResponseSnapshot:   response,
		ExtractedVariables: extracted,
		AssertionsResult:   assertionsResult,
		StartedAt:          row.StartedAt,
		FinishedAt:         row.FinishedAt,
		ResumeAt:           row.ResumeAt,
		Error:              row.Error,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}
