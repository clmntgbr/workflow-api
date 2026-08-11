package read

import (
	"context"
	"errors"
	"time"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type workflowRow struct {
	ID                      uuid.UUID
	Name                    string
	Description             string
	Status                  string
	OrganizationID          uuid.UUID
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func (workflowRow) TableName() string { return "workflows" }

type workflowReadRepository struct {
	db *gorm.DB
}

func NewWorkflowReadRepository(db *gorm.DB) domainworkflow.WorkflowReadRepository {
	return &workflowReadRepository{db: db}
}

func (r *workflowReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainworkflow.WorkflowView, error) {
	var row workflowRow
	err := r.db.WithContext(ctx).
		Select(
			"id", "name", "description", "status", "organization_id",
			"schedule_interval_minutes", "concurrency",
			"notifications_enabled", "notify_on_success", "notify_on_failure", "notify_on_cancel",
			"created_at", "updated_at",
		).
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
) ([]domainworkflow.WorkflowView, error) {
	var rows []workflowRow
	err := r.db.WithContext(ctx).
		Select(
			"id", "name", "description", "status", "organization_id",
			"schedule_interval_minutes", "concurrency",
			"notifications_enabled", "notify_on_success", "notify_on_failure", "notify_on_cancel",
			"created_at", "updated_at",
		).
		Where("organization_id = ? AND status <> ?", organizationID, domainworkflow.StatusDeleted).
		Order("created_at DESC").
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
	return &domainworkflow.WorkflowView{
		ID:                      row.ID,
		Name:                    row.Name,
		Description:             row.Description,
		Status:                  domainworkflow.Status(row.Status),
		OrganizationID:          row.OrganizationID,
		ScheduleIntervalMinutes: row.ScheduleIntervalMinutes,
		Concurrency:             row.Concurrency,
		NotificationsEnabled:    row.NotificationsEnabled,
		NotifyOnSuccess:         row.NotifyOnSuccess,
		NotifyOnFailure:         row.NotifyOnFailure,
		NotifyOnCancel:          row.NotifyOnCancel,
		CreatedAt:               row.CreatedAt,
		UpdatedAt:               row.UpdatedAt,
	}
}
