package read

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"go-api/internal/domain/paginate"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowRunRow struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	ProjectID    uuid.UUID
	Status            string
	TriggeredBy       string
	TriggeredByUserID *uuid.UUID
	Context           dbtype.JSONB
	StartedAt         *time.Time
	FinishedAt        *time.Time
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (workflowRunRow) TableName() string { return "workflow_runs" }

type workflowRunReadRepository struct {
	db *gorm.DB
}

func NewWorkflowRunReadRepository(db *gorm.DB) domainworkflowrun.WorkflowRunReadRepository {
	return &workflowRunReadRepository{db: db}
}

func (r *workflowRunReadRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainworkflowrun.WorkflowRunView, error) {
	var row workflowRunRow
	err := r.db.WithContext(ctx).
		Table("workflow_runs").
		Select(
			"workflow_runs.id, workflow_runs.workflow_id, workflows.project_id, workflow_runs.status, "+
				"workflow_runs.triggered_by, workflow_runs.triggered_by_user_id, workflow_runs.context, "+
				"workflow_runs.started_at, workflow_runs.finished_at, workflow_runs.error, "+
				"workflow_runs.created_at, workflow_runs.updated_at",
		).
		Joins("JOIN workflows ON workflows.id = workflow_runs.workflow_id").
		Where("workflow_runs.id = ?", id).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWorkflowRunView(row)
}

func (r *workflowRunReadRepository) FindByWorkflowID(
	ctx context.Context,
	workflowID uuid.UUID,
	query paginate.PaginateQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).
		Table("workflow_runs").
		Where("workflow_runs.workflow_id = ?", workflowID)

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []workflowRunRow
	err = db.
		Select(
			"workflow_runs.id, workflow_runs.workflow_id, workflows.project_id, workflow_runs.status, " +
				"workflow_runs.triggered_by, workflow_runs.triggered_by_user_id, workflow_runs.context, " +
				"workflow_runs.started_at, workflow_runs.finished_at, workflow_runs.error, " +
				"workflow_runs.created_at, workflow_runs.updated_at",
		).
		Joins("JOIN workflows ON workflows.id = workflow_runs.workflow_id").
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	views := make([]domainworkflowrun.WorkflowRunView, 0, len(rows))
	for _, row := range rows {
		view, err := toWorkflowRunView(row)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, *view)
	}
	return views, total, nil
}

func (r *workflowRunReadRepository) FindByProjectID(
	ctx context.Context,
	projectID uuid.UUID,
	query paginate.PaginateQuery,
) ([]domainworkflowrun.WorkflowRunView, int64, error) {
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "workflow_runs.created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).
		Table("workflow_runs").
		Joins("JOIN workflows ON workflows.id = workflow_runs.workflow_id").
		Where("workflows.project_id = ?", projectID)

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []workflowRunRow
	err = db.
		Select(
			"workflow_runs.id, workflow_runs.workflow_id, workflows.project_id, workflow_runs.status, " +
				"workflow_runs.triggered_by, workflow_runs.triggered_by_user_id, workflow_runs.context, " +
				"workflow_runs.started_at, workflow_runs.finished_at, workflow_runs.error, " +
				"workflow_runs.created_at, workflow_runs.updated_at",
		).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	views := make([]domainworkflowrun.WorkflowRunView, 0, len(rows))
	for _, row := range rows {
		view, err := toWorkflowRunView(row)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, *view)
	}
	return views, total, nil
}

func (r *workflowRunReadRepository) FindAnalyticsByProject(
	ctx context.Context,
	projectID uuid.UUID,
	filter domainworkflowrun.WorkflowRunAnalyticsFilter,
) (*domainworkflowrun.WorkflowRunAnalytics, error) {
	db := r.db.WithContext(ctx).
		Table("workflow_runs").
		Joins("JOIN workflows ON workflows.id = workflow_runs.workflow_id").
		Where("workflows.project_id = ?", projectID)

	if filter.WorkflowID != nil {
		db = db.Where("workflow_runs.workflow_id = ?", *filter.WorkflowID)
	}
	if filter.From != nil {
		db = db.Where("workflow_runs.created_at >= ?", *filter.From)
	}
	if filter.To != nil {
		db = db.Where("workflow_runs.created_at <= ?", *filter.To)
	}

	type analyticsRow struct {
		TotalRuns         int64
		SuccessCount      int64
		FailureCount      int64
		CancelledCount    int64
		RunningCount      int64
		PendingCount      int64
		AverageDurationMS sql.NullFloat64
		MinDurationMS     sql.NullFloat64
		MaxDurationMS     sql.NullFloat64
		LastRunAt         *time.Time
	}

	var row analyticsRow
	err := db.Select(`
		COUNT(*) AS total_runs,
		COUNT(*) FILTER (WHERE workflow_runs.status = 'success') AS success_count,
		COUNT(*) FILTER (WHERE workflow_runs.status = 'failed') AS failure_count,
		COUNT(*) FILTER (WHERE workflow_runs.status = 'cancelled') AS cancelled_count,
		COUNT(*) FILTER (WHERE workflow_runs.status = 'running') AS running_count,
		COUNT(*) FILTER (WHERE workflow_runs.status = 'pending') AS pending_count,
		AVG(EXTRACT(EPOCH FROM (workflow_runs.finished_at - workflow_runs.started_at)) * 1000)
			FILTER (WHERE workflow_runs.started_at IS NOT NULL AND workflow_runs.finished_at IS NOT NULL) AS average_duration_ms,
		MIN(EXTRACT(EPOCH FROM (workflow_runs.finished_at - workflow_runs.started_at)) * 1000)
			FILTER (WHERE workflow_runs.started_at IS NOT NULL AND workflow_runs.finished_at IS NOT NULL) AS min_duration_ms,
		MAX(EXTRACT(EPOCH FROM (workflow_runs.finished_at - workflow_runs.started_at)) * 1000)
			FILTER (WHERE workflow_runs.started_at IS NOT NULL AND workflow_runs.finished_at IS NOT NULL) AS max_duration_ms,
		MAX(workflow_runs.created_at) AS last_run_at
	`).Scan(&row).Error
	if err != nil {
		return nil, err
	}

	stats := &domainworkflowrun.WorkflowRunAnalytics{
		TotalRuns:         row.TotalRuns,
		SuccessCount:      row.SuccessCount,
		FailureCount:      row.FailureCount,
		CancelledCount:    row.CancelledCount,
		RunningCount:      row.RunningCount,
		PendingCount:      row.PendingCount,
		AverageDurationMS: 0,
		MinDurationMS:     0,
		MaxDurationMS:     0,
		LastRunAt:         row.LastRunAt,
	}
	if row.AverageDurationMS.Valid {
		stats.AverageDurationMS = row.AverageDurationMS.Float64
	}
	if row.MinDurationMS.Valid {
		stats.MinDurationMS = row.MinDurationMS.Float64
	}
	if row.MaxDurationMS.Valid {
		stats.MaxDurationMS = row.MaxDurationMS.Float64
	}
	return stats, nil
}

func toWorkflowRunView(row workflowRunRow) (*domainworkflowrun.WorkflowRunView, error) {
	ctx := map[string]any{}
	if len(row.Context) > 0 {
		if err := json.Unmarshal(row.Context, &ctx); err != nil {
			return nil, err
		}
	}

	return &domainworkflowrun.WorkflowRunView{
		ID:                row.ID,
		WorkflowID:        row.WorkflowID,
		ProjectID:    row.ProjectID,
		Status:            domainworkflowrun.Status(row.Status),
		TriggeredBy:       domainworkflowrun.TriggeredBy(row.TriggeredBy),
		TriggeredByUserID: row.TriggeredByUserID,
		Context:           ctx,
		StartedAt:         row.StartedAt,
		FinishedAt:        row.FinishedAt,
		Error:             row.Error,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}
