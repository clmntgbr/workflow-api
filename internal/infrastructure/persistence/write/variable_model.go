package write

import (
	"time"

	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type VariableModel struct {
	ID          uuid.UUID `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Key         string    `gorm:"column:key"`
	Description string    `gorm:"column:description"`
	Path        string    `gorm:"column:path"`
	StepID      uuid.UUID `gorm:"column:step_id"`
	WorkflowID  uuid.UUID `gorm:"column:workflow_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (VariableModel) TableName() string { return "variables" }

func variableModelFromDomain(v *domainvariable.Variable) *VariableModel {
	return &VariableModel{
		ID:         v.ID,
		Name:       v.Name,
		Key:        v.Key,
		Description: v.Description,
		Path:       v.Path,
		StepID:     v.StepID,
		WorkflowID: v.WorkflowID,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

func variableDomainFromModel(m *VariableModel) *domainvariable.Variable {
	return &domainvariable.Variable{
		ID:         m.ID,
		Name:       m.Name,
		Key:        m.Key,
		Description: m.Description,
		Path:       m.Path,
		StepID:     m.StepID,
		WorkflowID: m.WorkflowID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}
