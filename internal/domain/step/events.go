package step

import (
	"go-api/internal/domain/event"
	"time"

	"go-api/internal/domain/httpquery"
)

const EventTypeStepCreated = "step.created.v1"

type StepCreated struct {
	event.PerformedBy
	ID             string            `json:"eventId"`
	StepID         string            `json:"stepId"`
	WorkflowID     string            `json:"workflowId"`
	EndpointID           string            `json:"endpointId,omitempty"`
	ProjectID            string            `json:"projectId"`
	Type                 string            `json:"type"`
	DelayDurationSeconds int               `json:"delayDurationSeconds,omitempty"`
	Name                 string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          httpquery.Params  `json:"query"`
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
	event.PerformedBy
	ID             string            `json:"eventId"`
	StepID         string            `json:"stepId"`
	WorkflowID     string            `json:"workflowId"`
	ProjectID            string            `json:"projectId"`
	Type                 string            `json:"type"`
	DelayDurationSeconds int               `json:"delayDurationSeconds,omitempty"`
	Name                 string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          httpquery.Params  `json:"query"`
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

const EventTypeStepPositionUpdated = "step.position_updated.v1"

type StepPositionUpdated struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	StepID         string    `json:"stepId"`
	WorkflowID     string    `json:"workflowId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Index          string    `json:"index"`
	ExecutionOrder int       `json:"executionOrder"`
	TreeIndex      int       `json:"treeIndex"`
	Position       Position  `json:"position"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e StepPositionUpdated) EventID() string       { return e.ID }
func (e StepPositionUpdated) EventType() string     { return EventTypeStepPositionUpdated }
func (e StepPositionUpdated) AggregateID() string   { return e.StepID }
func (e StepPositionUpdated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeStepDeleted = "step.deleted.v1"

type StepDeleted struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	StepID         string    `json:"stepId"`
	WorkflowID     string    `json:"workflowId"`
	ProjectID string    `json:"projectId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e StepDeleted) EventID() string       { return e.ID }
func (e StepDeleted) EventType() string     { return EventTypeStepDeleted }
func (e StepDeleted) AggregateID() string   { return e.StepID }
func (e StepDeleted) OccurredAt() time.Time { return e.Timestamp }
