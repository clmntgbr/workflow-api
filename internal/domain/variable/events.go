package variable

import (
	"time"

	"go-api/internal/domain/event"
)

const EventTypeVariableCreated = "variable.created.v1"

type VariableCreated struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	VariableID     string    `json:"variableId"`
	WorkflowID     string    `json:"workflowId"`
	StepID         string    `json:"stepId,omitempty"`
	ProjectID string    `json:"projectId"`
	Name           string    `json:"name"`
	Key            string    `json:"key"`
	Description    string    `json:"description"`
	Kind           string    `json:"kind"`
	Path           string    `json:"path,omitempty"`
	Value          any       `json:"value,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e VariableCreated) EventID() string       { return e.ID }
func (e VariableCreated) EventType() string     { return EventTypeVariableCreated }
func (e VariableCreated) AggregateID() string   { return e.VariableID }
func (e VariableCreated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeVariableUpdated = "variable.updated.v1"

type VariableUpdated struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	VariableID     string    `json:"variableId"`
	WorkflowID     string    `json:"workflowId"`
	StepID         string    `json:"stepId,omitempty"`
	ProjectID string    `json:"projectId"`
	Name           string    `json:"name"`
	Key            string    `json:"key"`
	Description    string    `json:"description"`
	Kind           string    `json:"kind"`
	Path           string    `json:"path,omitempty"`
	Value          any       `json:"value,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e VariableUpdated) EventID() string       { return e.ID }
func (e VariableUpdated) EventType() string     { return EventTypeVariableUpdated }
func (e VariableUpdated) AggregateID() string   { return e.VariableID }
func (e VariableUpdated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeVariableDeleted = "variable.deleted.v1"

type VariableDeleted struct {
	event.PerformedBy
	ID         string    `json:"eventId"`
	VariableID string    `json:"variableId"`
	WorkflowID string    `json:"workflowId"`
	StepID     string    `json:"stepId,omitempty"`
	ProjectID  string    `json:"projectId"`
	Key        string    `json:"key"`
	Kind       string    `json:"kind"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e VariableDeleted) EventID() string       { return e.ID }
func (e VariableDeleted) EventType() string     { return EventTypeVariableDeleted }
func (e VariableDeleted) AggregateID() string   { return e.VariableID }
func (e VariableDeleted) OccurredAt() time.Time { return e.Timestamp }
