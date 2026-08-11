package write

import (
	"time"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

// WorkflowModel is the persistence mapping for table workflows.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type WorkflowModel struct {
	ID                      uuid.UUID `gorm:"column:id;primaryKey"`
	Name                    string    `gorm:"column:name"`
	Description             string    `gorm:"column:description"`
	Status                  string    `gorm:"column:status"`
	OrganizationID          uuid.UUID `gorm:"column:organization_id"`
	ScheduleIntervalMinutes int       `gorm:"column:schedule_interval_minutes"`
	Concurrency             int       `gorm:"column:concurrency"`
	NotificationsEnabled    bool      `gorm:"column:notifications_enabled"`
	NotifyOnSuccess         bool      `gorm:"column:notify_on_success"`
	NotifyOnFailure         bool      `gorm:"column:notify_on_failure"`
	NotifyOnCancel          bool      `gorm:"column:notify_on_cancel"`
	CreatedAt               time.Time `gorm:"column:created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at"`
}

func (WorkflowModel) TableName() string {
	return "workflows"
}

func workflowModelFromDomain(w *domainworkflow.Workflow) *WorkflowModel {
	return &WorkflowModel{
		ID:                      w.ID,
		Name:                    w.Name,
		Description:             w.Description,
		Status:                  string(w.Status),
		OrganizationID:          w.OrganizationID,
		ScheduleIntervalMinutes: w.ScheduleIntervalMinutes,
		Concurrency:             w.Concurrency,
		NotificationsEnabled:    w.NotificationsEnabled,
		NotifyOnSuccess:         w.NotifyOnSuccess,
		NotifyOnFailure:         w.NotifyOnFailure,
		NotifyOnCancel:          w.NotifyOnCancel,
		CreatedAt:               w.CreatedAt,
		UpdatedAt:               w.UpdatedAt,
	}
}

func workflowDomainFromModel(m *WorkflowModel) *domainworkflow.Workflow {
	return &domainworkflow.Workflow{
		ID:                      m.ID,
		Name:                    m.Name,
		Description:             m.Description,
		Status:                  domainworkflow.Status(m.Status),
		OrganizationID:          m.OrganizationID,
		ScheduleIntervalMinutes: m.ScheduleIntervalMinutes,
		Concurrency:             m.Concurrency,
		NotificationsEnabled:    m.NotificationsEnabled,
		NotifyOnSuccess:         m.NotifyOnSuccess,
		NotifyOnFailure:         m.NotifyOnFailure,
		NotifyOnCancel:          m.NotifyOnCancel,
		CreatedAt:               m.CreatedAt,
		UpdatedAt:               m.UpdatedAt,
	}
}
