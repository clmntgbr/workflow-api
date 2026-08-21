package write

import (
	"encoding/json"
	"time"

	"go-api/internal/infrastructure/persistence/dbtype"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type VariableModel struct {
	ID          uuid.UUID   `gorm:"column:id;primaryKey"`
	Name        string      `gorm:"column:name"`
	Key         string      `gorm:"column:key"`
	Description string      `gorm:"column:description"`
	Kind        string      `gorm:"column:kind"`
	Path        *string     `gorm:"column:path"`
	Value       dbtype.JSONB `gorm:"column:value;type:jsonb"`
	StepID      *uuid.UUID  `gorm:"column:step_id"`
	WorkflowID  uuid.UUID   `gorm:"column:workflow_id"`
	CreatedAt   time.Time   `gorm:"column:created_at"`
	UpdatedAt   time.Time   `gorm:"column:updated_at"`
}

func (VariableModel) TableName() string { return "variables" }

func variableModelFromDomain(v *domainvariable.Variable) *VariableModel {
	var path *string
	if v.Kind == domainvariable.KindExtracted {
		p := v.Path
		path = &p
	}

	var valueRaw dbtype.JSONB
	if v.Kind == domainvariable.KindStatic && v.Value != nil {
		encoded, err := json.Marshal(v.Value)
		if err == nil {
			valueRaw = dbtype.JSONB(encoded)
		}
	}

	return &VariableModel{
		ID:          v.ID,
		Name:        v.Name,
		Key:         v.Key,
		Description: v.Description,
		Kind:        string(v.Kind),
		Path:        path,
		Value:       valueRaw,
		StepID:      v.StepID,
		WorkflowID:  v.WorkflowID,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func variableDomainFromModel(m *VariableModel) *domainvariable.Variable {
	kind := domainvariable.Kind(m.Kind)
	if kind == "" {
		kind = domainvariable.KindExtracted
	}

	path := ""
	if m.Path != nil {
		path = *m.Path
	}

	var value any
	if len(m.Value) > 0 {
		_ = json.Unmarshal(m.Value, &value)
	}

	return &domainvariable.Variable{
		ID:          m.ID,
		Name:        m.Name,
		Key:         m.Key,
		Description: m.Description,
		Kind:        kind,
		Path:        path,
		Value:       value,
		StepID:      m.StepID,
		WorkflowID:  m.WorkflowID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
