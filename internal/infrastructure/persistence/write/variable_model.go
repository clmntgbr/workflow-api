package write

import (
	"encoding/json"
	"time"

	domainvariable "go-api/internal/domain/variable"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

type VariableModel struct {
	ID           uuid.UUID    `gorm:"column:id;primaryKey"`
	Name         string       `gorm:"column:name"`
	Key          string       `gorm:"column:key"`
	Description  string       `gorm:"column:description"`
	Path         string       `gorm:"column:path"`
	StepID       uuid.UUID    `gorm:"column:step_id"`
	WorkflowID   uuid.UUID    `gorm:"column:workflow_id"`
	IsSecret     bool         `gorm:"column:is_secret"`
	DefaultValue dbtype.JSONB `gorm:"column:default_value"`
	LastValue    dbtype.JSONB `gorm:"column:last_value"`
	CreatedAt    time.Time    `gorm:"column:created_at"`
	UpdatedAt    time.Time    `gorm:"column:updated_at"`
}

func (VariableModel) TableName() string { return "variables" }

func variableModelFromDomain(v *domainvariable.Variable) *VariableModel {
	return &VariableModel{
		ID:           v.ID,
		Name:         v.Name,
		Key:          v.Key,
		Description:  v.Description,
		Path:         v.Path,
		StepID:       v.StepID,
		WorkflowID:   v.WorkflowID,
		IsSecret:     v.IsSecret,
		DefaultValue: dbtype.JSONB(v.DefaultValue),
		LastValue:    dbtype.JSONB(v.LastValue),
		CreatedAt:    v.CreatedAt,
		UpdatedAt:    v.UpdatedAt,
	}
}

func variableDomainFromModel(m *VariableModel) *domainvariable.Variable {
	return &domainvariable.Variable{
		ID:           m.ID,
		Name:         m.Name,
		Key:          m.Key,
		Description:  m.Description,
		Path:         m.Path,
		StepID:       m.StepID,
		WorkflowID:   m.WorkflowID,
		IsSecret:     m.IsSecret,
		DefaultValue: json.RawMessage(m.DefaultValue),
		LastValue:    json.RawMessage(m.LastValue),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
