package write

import (
	domainconnection "go-api/internal/domain/connection"

	"github.com/google/uuid"
)

type ConnectionModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey"`
	WorkflowID   uuid.UUID `gorm:"column:workflow_id"`
	ProjectID    uuid.UUID `gorm:"column:project_id"`
	SourceStepID uuid.UUID `gorm:"column:source_step_id"`
	TargetStepID uuid.UUID `gorm:"column:target_step_id"`
	Branch       *string   `gorm:"column:branch"`
}

func (ConnectionModel) TableName() string {
	return "connections"
}

func connectionModelFromDomain(c *domainconnection.Connection) *ConnectionModel {
	return &ConnectionModel{
		ID:           c.ID,
		WorkflowID:   c.WorkflowID,
		ProjectID:    c.ProjectID,
		SourceStepID: c.SourceStepID,
		TargetStepID: c.TargetStepID,
		Branch:       branchPtrFromDomain(c.Branch),
	}
}

func connectionDomainFromModel(m *ConnectionModel) *domainconnection.Connection {
	return &domainconnection.Connection{
		ID:           m.ID,
		WorkflowID:   m.WorkflowID,
		ProjectID:    m.ProjectID,
		SourceStepID: m.SourceStepID,
		TargetStepID: m.TargetStepID,
		Branch:       branchFromModel(m.Branch),
	}
}

func branchPtrFromDomain(branch *domainconnection.ConditionBranch) *string {
	if branch == nil {
		return nil
	}
	value := string(*branch)
	return &value
}

func branchFromModel(branch *string) *domainconnection.ConditionBranch {
	if branch == nil {
		return nil
	}
	value := domainconnection.ConditionBranch(*branch)
	return &value
}
