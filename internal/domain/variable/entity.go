package variable

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Variable struct {
	ID          uuid.UUID
	Name        string
	Key         string
	Description string
	Kind        Kind
	Path        string
	Value       any
	StepID      *uuid.UUID
	WorkflowID  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time

	events []event.DomainEvent
}

type NewVariableParams struct {
	Name           string
	Key            string
	Description    string
	Kind           Kind
	Path           string
	Value          any
	StepID         *uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
}

func NewVariable(p NewVariableParams) (*Variable, error) {
	if err := ValidateShape(p.Kind, p.StepID, p.Path, p.Value); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	v := &Variable{
		ID:          uuid.New(),
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		Kind:        p.Kind,
		Path:        normalizePath(p.Kind, p.Path),
		Value:       normalizeValue(p.Kind, p.Value),
		StepID:      p.StepID,
		WorkflowID:  p.WorkflowID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	v.recordEvent(VariableCreated{
		ID:             uuid.New().String(),
		VariableID:     v.ID.String(),
		WorkflowID:     v.WorkflowID.String(),
		StepID:         optionalUUIDString(v.StepID),
		ProjectID: p.ProjectID.String(),
		Name:           v.Name,
		Key:            v.Key,
		Description:    v.Description,
		Kind:           string(v.Kind),
		Path:           v.Path,
		Value:          v.Value,
		Timestamp:      now,
	})
	return v, nil
}

func (v *Variable) Update(p UpdateVariableParams) error {
	path := p.Path
	value := p.Value
	if v.Kind == KindStatic {
		path = ""
		value = p.Value
	} else {
		value = nil
	}
	if err := ValidateShape(v.Kind, v.StepID, path, value); err != nil {
		return err
	}

	v.Name = p.Name
	v.Key = p.Key
	v.Description = p.Description
	v.Path = normalizePath(v.Kind, path)
	v.Value = normalizeValue(v.Kind, value)
	v.UpdatedAt = time.Now().UTC()
	v.recordEvent(VariableUpdated{
		ID:             uuid.New().String(),
		VariableID:     v.ID.String(),
		WorkflowID:     v.WorkflowID.String(),
		StepID:         optionalUUIDString(v.StepID),
		ProjectID: p.ProjectID.String(),
		Name:           v.Name,
		Key:            v.Key,
		Description:    v.Description,
		Kind:           string(v.Kind),
		Path:           v.Path,
		Value:          v.Value,
		Timestamp:      v.UpdatedAt,
	})
	return nil
}

type UpdateVariableParams struct {
	Name           string
	Key            string
	Description    string
	Path           string
	Value          any
	ProjectID uuid.UUID
}

func (v *Variable) PullEvents() []event.DomainEvent {
	events := v.events
	v.events = nil
	return events
}

func (v *Variable) recordEvent(evt event.DomainEvent) {
	v.events = append(v.events, evt)
}

func (v *Variable) IsStatic() bool {
	return v.Kind == KindStatic
}

func optionalUUIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func normalizePath(kind Kind, path string) string {
	if kind == KindStatic {
		return ""
	}
	return path
}

func normalizeValue(kind Kind, value any) any {
	if kind == KindExtracted {
		return nil
	}
	return value
}
