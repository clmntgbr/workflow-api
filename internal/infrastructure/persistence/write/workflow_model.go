package write

import (
	"time"

	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type WorkflowModel struct {
	ID                    uuid.UUID  `gorm:"column:id;primaryKey"`
	Name                  string     `gorm:"column:name"`
	Description           string     `gorm:"column:description"`
	Status                string     `gorm:"column:status"`
	ProjectID        uuid.UUID  `gorm:"column:project_id"`
	ScheduleType          string     `gorm:"column:schedule_type"`
	ScheduleIntervalValue int        `gorm:"column:schedule_interval_value"`
	ScheduleIntervalUnit  *string    `gorm:"column:schedule_interval_unit"`
	ScheduleAt            *time.Time `gorm:"column:schedule_at"`
	ScheduleTimezone      string     `gorm:"column:schedule_timezone"`
	NextRunAt             *time.Time `gorm:"column:next_run_at"`
	Concurrency           int        `gorm:"column:concurrency"`
	NotificationsEnabled  bool       `gorm:"column:notifications_enabled"`
	NotifyOnSuccess       bool       `gorm:"column:notify_on_success"`
	NotifyOnFailure       bool       `gorm:"column:notify_on_failure"`
	NotifyOnCancel        bool       `gorm:"column:notify_on_cancel"`
	CreatedAt             time.Time  `gorm:"column:created_at"`
	UpdatedAt             time.Time  `gorm:"column:updated_at"`
}

func (WorkflowModel) TableName() string {
	return "workflows"
}

func workflowModelFromDomain(w *domainworkflow.Workflow) *WorkflowModel {
	return &WorkflowModel{
		ID:                    w.ID,
		Name:                  w.Name,
		Description:           w.Description,
		Status:                string(w.Status),
		ProjectID:        w.ProjectID,
		ScheduleType:          string(w.ScheduleType),
		ScheduleIntervalValue: w.ScheduleIntervalValue,
		ScheduleIntervalUnit:  scheduleUnitPointer(w.ScheduleIntervalUnit),
		ScheduleAt:            w.ScheduleAt,
		ScheduleTimezone:      w.ScheduleTimezone,
		NextRunAt:             w.NextRunAt,
		Concurrency:           w.Concurrency,
		NotificationsEnabled:  w.NotificationsEnabled,
		NotifyOnSuccess:       w.NotifyOnSuccess,
		NotifyOnFailure:       w.NotifyOnFailure,
		NotifyOnCancel:        w.NotifyOnCancel,
		CreatedAt:             w.CreatedAt,
		UpdatedAt:             w.UpdatedAt,
	}
}

func workflowDomainFromModel(m *WorkflowModel) *domainworkflow.Workflow {
	return &domainworkflow.Workflow{
		ID:                    m.ID,
		Name:                  m.Name,
		Description:           m.Description,
		Status:                domainworkflow.Status(m.Status),
		ProjectID:        m.ProjectID,
		ScheduleType:          domainworkflow.ScheduleType(m.ScheduleType),
		ScheduleIntervalValue: m.ScheduleIntervalValue,
		ScheduleIntervalUnit:  scheduleUnitFromPointer(m.ScheduleIntervalUnit),
		ScheduleAt:            m.ScheduleAt,
		ScheduleTimezone:      m.ScheduleTimezone,
		NextRunAt:             m.NextRunAt,
		Concurrency:           m.Concurrency,
		NotificationsEnabled:  m.NotificationsEnabled,
		NotifyOnSuccess:       m.NotifyOnSuccess,
		NotifyOnFailure:       m.NotifyOnFailure,
		NotifyOnCancel:        m.NotifyOnCancel,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}

func scheduleUnitPointer(unit domainworkflow.ScheduleUnit) *string {
	if unit == "" {
		return nil
	}
	value := string(unit)
	return &value
}

func scheduleUnitFromPointer(value *string) domainworkflow.ScheduleUnit {
	if value == nil {
		return ""
	}
	return domainworkflow.ScheduleUnit(*value)
}
