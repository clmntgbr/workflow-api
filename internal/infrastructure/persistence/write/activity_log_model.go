package write

import (
	"time"

	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
)

type ActivityLogModel struct {
	ID              uuid.UUID  `gorm:"column:id;primaryKey"`
	ProjectID       uuid.UUID  `gorm:"column:project_id"`
	Action            string     `gorm:"column:action"`
	SubjectType       string     `gorm:"column:subject_type"`
	SubjectID         uuid.UUID  `gorm:"column:subject_id"`
	WorkflowID        *uuid.UUID `gorm:"column:workflow_id"`
	WorkflowRunID     *uuid.UUID `gorm:"column:workflow_run_id"`
	StepID            *uuid.UUID `gorm:"column:step_id"`
	StepRunID         *uuid.UUID `gorm:"column:step_run_id"`
	ActorType         string     `gorm:"column:actor_type"`
	ActorUserID       *uuid.UUID `gorm:"column:actor_user_id"`
	Level             string     `gorm:"column:level"`
	Message           string     `gorm:"column:message"`
	SourceEventID     uuid.UUID  `gorm:"column:source_event_id"`
	SourceEventType   string     `gorm:"column:source_event_type"`
	OccurredAt        time.Time  `gorm:"column:occurred_at"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
}

func (ActivityLogModel) TableName() string {
	return "activity_logs"
}

func activityLogModelFromDomain(e *domainactivitylog.Entry) *ActivityLogModel {
	return &ActivityLogModel{
		ID:              e.ID,
		ProjectID:       e.ProjectID,
		Action:          e.Action,
		SubjectType:     e.SubjectType,
		SubjectID:       e.SubjectID,
		WorkflowID:      e.WorkflowID,
		WorkflowRunID:   e.WorkflowRunID,
		StepID:          e.StepID,
		StepRunID:       e.StepRunID,
		ActorType:       e.ActorType,
		ActorUserID:     e.ActorUserID,
		Level:           e.Level,
		Message:         e.Message,
		SourceEventID:   e.SourceEventID,
		SourceEventType: e.SourceEventType,
		OccurredAt:      e.OccurredAt,
		CreatedAt:       e.CreatedAt,
	}
}
