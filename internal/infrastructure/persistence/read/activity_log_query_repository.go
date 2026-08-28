package read

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/domain/paginate"
	domainactivitylog "go-api/internal/domain/activitylog"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type activityLogRow struct {
	ID              uuid.UUID
	ProjectID       uuid.UUID
	Action          string
	SubjectType     string
	SubjectID       uuid.UUID
	WorkflowID      *uuid.UUID
	WorkflowRunID   *uuid.UUID
	StepID          *uuid.UUID
	StepRunID       *uuid.UUID
	ActorType       string
	ActorUserID     *uuid.UUID
	Level           string
	Message         string
	Payload         dbtype.JSONB
	SourceEventID   uuid.UUID
	SourceEventType string
	OccurredAt      time.Time
	CreatedAt       time.Time
}

func (activityLogRow) TableName() string {
	return "activity_logs"
}

type activityLogReadRepository struct {
	db *gorm.DB
}

func NewActivityLogReadRepository(db *gorm.DB) domainactivitylog.ReadRepository {
	return &activityLogReadRepository{db: db}
}

func (r *activityLogReadRepository) FindByWorkflowID(
	ctx context.Context,
	workflowID uuid.UUID,
	query paginate.PaginateQuery,
) ([]domainactivitylog.View, int64, error) {
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "occurred_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	base := r.db.WithContext(ctx).
		Table("activity_logs").
		Where(
			"activity_logs.workflow_id = ? OR activity_logs.workflow_run_id IN ("+
				"SELECT id FROM workflow_runs WHERE workflow_id = ?)",
			workflowID,
			workflowID,
		)

	paginated, total, err := Paginate(base, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []activityLogRow
	if err := paginated.Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	views := make([]domainactivitylog.View, 0, len(rows))
	for _, row := range rows {
		view, err := toActivityLogView(row)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, view)
	}
	return views, total, nil
}

func toActivityLogView(row activityLogRow) (domainactivitylog.View, error) {
	payload := map[string]any{}
	if len(row.Payload) > 0 {
		if err := json.Unmarshal(row.Payload, &payload); err != nil {
			return domainactivitylog.View{}, err
		}
	}

	return domainactivitylog.View{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		Action:          row.Action,
		SubjectType:     row.SubjectType,
		SubjectID:       row.SubjectID,
		WorkflowID:      row.WorkflowID,
		WorkflowRunID:   row.WorkflowRunID,
		StepID:          row.StepID,
		StepRunID:       row.StepRunID,
		ActorType:       row.ActorType,
		ActorUserID:     row.ActorUserID,
		Level:           row.Level,
		Message:         row.Message,
		Payload:         payload,
		SourceEventID:   row.SourceEventID,
		SourceEventType: row.SourceEventType,
		OccurredAt:      row.OccurredAt,
		CreatedAt:       row.CreatedAt,
	}, nil
}
