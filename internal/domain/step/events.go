package step

import "time"

const EventTypeStepCreated = "step.created.v1"

type StepCreated struct {
	ID             string            `json:"eventId"`
	StepID         string            `json:"stepId"`
	WorkflowID     string            `json:"workflowId"`
	EndpointID     string            `json:"endpointId"`
	OrganizationID string            `json:"organizationId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Index          string            `json:"index"`
	ExecutionOrder int               `json:"executionOrder"`
	TreeIndex      int               `json:"treeIndex"`
	Position       Position          `json:"position"`
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e StepCreated) EventID() string       { return e.ID }
func (e StepCreated) EventType() string     { return EventTypeStepCreated }
func (e StepCreated) AggregateID() string   { return e.StepID }
func (e StepCreated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeStepUpdated = "step.updated.v1"

type StepUpdated struct {
	ID             string            `json:"eventId"`
	StepID         string            `json:"stepId"`
	WorkflowID     string            `json:"workflowId"`
	OrganizationID string            `json:"organizationId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Index          string            `json:"index"`
	ExecutionOrder int               `json:"executionOrder"`
	TreeIndex      int               `json:"treeIndex"`
	Position       Position          `json:"position"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e StepUpdated) EventID() string       { return e.ID }
func (e StepUpdated) EventType() string     { return EventTypeStepUpdated }
func (e StepUpdated) AggregateID() string   { return e.StepID }
func (e StepUpdated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeStepDeleted = "step.deleted.v1"

type StepDeleted struct {
	ID             string    `json:"eventId"`
	StepID         string    `json:"stepId"`
	WorkflowID     string    `json:"workflowId"`
	OrganizationID string    `json:"organizationId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e StepDeleted) EventID() string       { return e.ID }
func (e StepDeleted) EventType() string     { return EventTypeStepDeleted }
func (e StepDeleted) AggregateID() string   { return e.StepID }
func (e StepDeleted) OccurredAt() time.Time { return e.Timestamp }
