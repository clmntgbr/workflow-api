package write

import (
	"time"

	domainrunexport "go-api/internal/domain/runexport"

	"github.com/google/uuid"
)

type RunExportModel struct {
	ID                uuid.UUID `gorm:"column:id;primaryKey"`
	WorkflowID        uuid.UUID `gorm:"column:workflow_id"`
	ProjectID         uuid.UUID `gorm:"column:project_id"`
	RequestedByUserID uuid.UUID `gorm:"column:requested_by_user_id"`
	DateRangeFrom     time.Time `gorm:"column:date_range_from"`
	DateRangeTo       time.Time `gorm:"column:date_range_to"`
	Status            string    `gorm:"column:status"`
	Error             string    `gorm:"column:error"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (RunExportModel) TableName() string {
	return "run_exports"
}

func runExportModelFromDomain(j *domainrunexport.RunExport) *RunExportModel {
	return &RunExportModel{
		ID:                j.ID,
		WorkflowID:        j.WorkflowID,
		ProjectID:         j.ProjectID,
		RequestedByUserID: j.RequestedByUserID,
		DateRangeFrom:     j.DateRangeFrom,
		DateRangeTo:       j.DateRangeTo,
		Status:            string(j.Status),
		Error:             j.Error,
		CreatedAt:         j.CreatedAt,
		UpdatedAt:         j.UpdatedAt,
	}
}

func runExportDomainFromModel(m *RunExportModel) *domainrunexport.RunExport {
	return &domainrunexport.RunExport{
		ID:                m.ID,
		WorkflowID:        m.WorkflowID,
		ProjectID:         m.ProjectID,
		RequestedByUserID: m.RequestedByUserID,
		DateRangeFrom:     m.DateRangeFrom,
		DateRangeTo:       m.DateRangeTo,
		Status:            domainrunexport.Status(m.Status),
		Error:             m.Error,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}
