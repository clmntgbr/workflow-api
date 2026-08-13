package read

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/paginate"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowRow struct {
	ID                    uuid.UUID
	Name                  string
	Description           string
	Status                string
	OrganizationID        uuid.UUID
	ScheduleType          string
	ScheduleIntervalValue int
	ScheduleIntervalUnit  *string
	ScheduleAt            *time.Time
	ScheduleTimezone      string
	NextRunAt             *time.Time
	Concurrency           int
	NotificationsEnabled  bool
	NotifyOnSuccess       bool
	NotifyOnFailure       bool
	NotifyOnCancel        bool
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (workflowRow) TableName() string { return "workflows" }

var workflowSelectColumns = []string{
	"id", "name", "description", "status", "organization_id",
	"schedule_type", "schedule_interval_value", "schedule_interval_unit",
	"schedule_at", "schedule_timezone", "next_run_at",
	"concurrency",
	"notifications_enabled", "notify_on_success", "notify_on_failure", "notify_on_cancel",
	"created_at", "updated_at",
}

type workflowReadRepository struct {
	db *gorm.DB
}

func NewWorkflowReadRepository(db *gorm.DB) domainworkflow.WorkflowReadRepository {
	return &workflowReadRepository{db: db}
}

func (r *workflowReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainworkflow.WorkflowView, error) {
	var row workflowRow
	err := r.db.WithContext(ctx).
		Select(workflowSelectColumns).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWorkflowView(row), nil
}

func (r *workflowReadRepository) FindByOrganizationID(
	ctx context.Context,
	organizationID uuid.UUID,
	query paginate.PaginateQuery,
) ([]domainworkflow.WorkflowView, int64, error) {
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).
		Model(&workflowRow{}).
		Where("organization_id = ? AND status <> ?", organizationID, domainworkflow.StatusDeleted)

	if query.Search != "" {
		like := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []workflowRow
	err = db.Select(workflowSelectColumns).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	views := make([]domainworkflow.WorkflowView, 0, len(rows))
	for _, row := range rows {
		views = append(views, *toWorkflowView(row))
	}
	return views, total, nil
}

func (r *workflowReadRepository) GetWorkflowsForExecution(ctx context.Context) ([]domainworkflow.WorkflowView, error) {
	var rows []workflowRow
	err := r.db.WithContext(ctx).
		Model(&workflowRow{}).
		Where("status = ?", domainworkflow.StatusActive).
		Where("next_run_at IS NOT NULL AND next_run_at <= ?", time.Now().UTC()).
		Where("EXISTS (SELECT 1 FROM steps WHERE steps.workflow_id = workflows.id AND steps.status <> ?)", domainstep.StatusDeleted).
		Order("next_run_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domainworkflow.WorkflowView, 0, len(rows))
	for _, row := range rows {
		views = append(views, *toWorkflowView(row))
	}
	return views, nil
}

func toWorkflowView(row workflowRow) *domainworkflow.WorkflowView {
	unit := domainworkflow.ScheduleUnit("")
	if row.ScheduleIntervalUnit != nil {
		unit = domainworkflow.ScheduleUnit(*row.ScheduleIntervalUnit)
	}
	timezone := row.ScheduleTimezone
	if timezone == "" {
		timezone = "UTC"
	}

	return &domainworkflow.WorkflowView{
		ID:                    row.ID,
		Name:                  row.Name,
		Description:           row.Description,
		Status:                domainworkflow.Status(row.Status),
		OrganizationID:        row.OrganizationID,
		ScheduleType:          domainworkflow.ScheduleType(row.ScheduleType),
		ScheduleIntervalValue: row.ScheduleIntervalValue,
		ScheduleIntervalUnit:  unit,
		ScheduleAt:            row.ScheduleAt,
		ScheduleTimezone:      timezone,
		NextRunAt:             row.NextRunAt,
		Concurrency:           row.Concurrency,
		NotificationsEnabled:  row.NotificationsEnabled,
		NotifyOnSuccess:       row.NotifyOnSuccess,
		NotifyOnFailure:       row.NotifyOnFailure,
		NotifyOnCancel:        row.NotifyOnCancel,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
	}
}
