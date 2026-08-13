package read

import (
	"context"
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
	OrganizationID    uuid.UUID
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
			"workflow_runs.id, workflow_runs.workflow_id, workflows.organization_id, workflow_runs.status, "+
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
			"workflow_runs.id, workflow_runs.workflow_id, workflows.organization_id, workflow_runs.status, " +
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
		OrganizationID:    row.OrganizationID,
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
