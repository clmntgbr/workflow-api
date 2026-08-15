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
	Path        string
	StepID      uuid.UUID
	WorkflowID  uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time

	events []event.DomainEvent
}

type NewVariableParams struct {
	Name           string
	Key            string
	Description    string
	Path           string
	StepID         uuid.UUID
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
}

func NewVariable(p NewVariableParams) *Variable {
	now := time.Now().UTC()
	v := &Variable{
		ID:          uuid.New(),
		Name:        p.Name,
		Key:         p.Key,
		Description: p.Description,
		Path:        p.Path,
		StepID:      p.StepID,
		WorkflowID:  p.WorkflowID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	v.recordEvent(VariableCreated{
		ID:             uuid.New().String(),
		VariableID:     v.ID.String(),
		WorkflowID:     v.WorkflowID.String(),
		StepID:         v.StepID.String(),
		OrganizationID: p.OrganizationID.String(),
		Name:           v.Name,
		Key:            v.Key,
		Description:    v.Description,
		Path:           v.Path,
		Timestamp:      now,
	})
	return v
}

func (v *Variable) Update(p UpdateVariableParams) {
	v.Name = p.Name
	v.Key = p.Key
	v.Description = p.Description
	v.Path = p.Path
	v.UpdatedAt = time.Now().UTC()
	v.recordEvent(VariableUpdated{
		ID:             uuid.New().String(),
		VariableID:     v.ID.String(),
		WorkflowID:     v.WorkflowID.String(),
		StepID:         v.StepID.String(),
		OrganizationID: p.OrganizationID.String(),
		Name:           v.Name,
		Key:            v.Key,
		Description:    v.Description,
		Path:           v.Path,
		Timestamp:      v.UpdatedAt,
	})
}

type UpdateVariableParams struct {
	Name           string
	Key            string
	Description    string
	Path           string
	OrganizationID uuid.UUID
}

func (v *Variable) PullEvents() []event.DomainEvent {
	events := v.events
	v.events = nil
	return events
}

func (v *Variable) recordEvent(evt event.DomainEvent) {
	v.events = append(v.events, evt)
}
