package variable

import "time"

const EventTypeVariableCreated = "variable.created.v1"

type VariableCreated struct {
	ID             string    `json:"eventId"`
	VariableID     string    `json:"variableId"`
	WorkflowID     string    `json:"workflowId"`
	StepID         string    `json:"stepId"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Key            string    `json:"key"`
	Description    string    `json:"description"`
	Path           string    `json:"path"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e VariableCreated) EventID() string       { return e.ID }
func (e VariableCreated) EventType() string     { return EventTypeVariableCreated }
func (e VariableCreated) AggregateID() string   { return e.VariableID }
func (e VariableCreated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeVariableUpdated = "variable.updated.v1"

type VariableUpdated struct {
	ID             string    `json:"eventId"`
	VariableID     string    `json:"variableId"`
	WorkflowID     string    `json:"workflowId"`
	StepID         string    `json:"stepId"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Key            string    `json:"key"`
	Description    string    `json:"description"`
	Path           string    `json:"path"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e VariableUpdated) EventID() string       { return e.ID }
func (e VariableUpdated) EventType() string     { return EventTypeVariableUpdated }
func (e VariableUpdated) AggregateID() string   { return e.VariableID }
func (e VariableUpdated) OccurredAt() time.Time { return e.Timestamp }
