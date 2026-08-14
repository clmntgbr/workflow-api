package variable

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Variable struct {
	ID           uuid.UUID
	Name         string
	Key          string
	Description  string
	Path         string
	StepID       uuid.UUID
	WorkflowID   uuid.UUID
	IsSecret     bool
	DefaultValue json.RawMessage
	LastValue    json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NewVariableParams struct {
	Name         string
	Key          string
	Description  string
	Path         string
	StepID       uuid.UUID
	WorkflowID   uuid.UUID
	IsSecret     bool
	DefaultValue json.RawMessage
}

func NewVariable(p NewVariableParams) *Variable {
	now := time.Now().UTC()
	return &Variable{
		ID:           uuid.New(),
		Name:         p.Name,
		Key:          p.Key,
		Description:  p.Description,
		Path:         p.Path,
		StepID:       p.StepID,
		WorkflowID:   p.WorkflowID,
		IsSecret:     p.IsSecret,
		DefaultValue: cloneRaw(p.DefaultValue),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (v *Variable) Update(p UpdateVariableParams) {
	v.Name = p.Name
	v.Key = p.Key
	v.Description = p.Description
	v.Path = p.Path
	v.IsSecret = p.IsSecret
	v.DefaultValue = cloneRaw(p.DefaultValue)
	if p.IsSecret {
		v.LastValue = nil
	}
	v.UpdatedAt = time.Now().UTC()
}

type UpdateVariableParams struct {
	Name         string
	Key          string
	Description  string
	Path         string
	IsSecret     bool
	DefaultValue json.RawMessage
}

func (v *Variable) SetLastValue(value json.RawMessage) {
	if v.IsSecret {
		v.LastValue = nil
		v.UpdatedAt = time.Now().UTC()
		return
	}
	v.LastValue = cloneRaw(value)
	v.UpdatedAt = time.Now().UTC()
}

func cloneRaw(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	out := make(json.RawMessage, len(value))
	copy(out, value)
	return out
}
