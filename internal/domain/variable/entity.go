package variable

import (
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
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type NewVariableParams struct {
	Name        string
	Key         string
	Description string
	Path        string
	StepID      uuid.UUID
	WorkflowID  uuid.UUID
}

func NewVariable(p NewVariableParams) *Variable {
	now := time.Now().UTC()
	return &Variable{
		ID:         uuid.New(),
		Name:       p.Name,
		Key:        p.Key,
		Description: p.Description,
		Path:       p.Path,
		StepID:     p.StepID,
		WorkflowID: p.WorkflowID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func (v *Variable) Update(p UpdateVariableParams) {
	v.Name = p.Name
	v.Key = p.Key
	v.Description = p.Description
	v.Path = p.Path
	v.UpdatedAt = time.Now().UTC()
}

type UpdateVariableParams struct {
	Name        string
	Key         string
	Description string
	Path        string
}
