package assertion

import (
	"time"

	"go-api/internal/domain/event"
)

const EventTypeAssertionCreated = "assertion.created.v1"

type AssertionCreated struct {
	event.PerformedBy
	ID            string    `json:"eventId"`
	AssertionID   string    `json:"assertionId"`
	WorkflowID    string    `json:"workflowId"`
	StepID        string    `json:"stepId"`
	ProjectID     string    `json:"projectId"`
	Description   string    `json:"description"`
	Source        string    `json:"source"`
	Path          string    `json:"path,omitempty"`
	Operator      string    `json:"operator"`
	ExpectedValue string    `json:"expectedValue,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e AssertionCreated) EventID() string       { return e.ID }
func (e AssertionCreated) EventType() string     { return EventTypeAssertionCreated }
func (e AssertionCreated) AggregateID() string   { return e.AssertionID }
func (e AssertionCreated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeAssertionUpdated = "assertion.updated.v1"

type AssertionUpdated struct {
	event.PerformedBy
	ID            string    `json:"eventId"`
	AssertionID   string    `json:"assertionId"`
	WorkflowID    string    `json:"workflowId"`
	StepID        string    `json:"stepId"`
	ProjectID     string    `json:"projectId"`
	Description   string    `json:"description"`
	Source        string    `json:"source"`
	Path          string    `json:"path,omitempty"`
	Operator      string    `json:"operator"`
	ExpectedValue string    `json:"expectedValue,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e AssertionUpdated) EventID() string       { return e.ID }
func (e AssertionUpdated) EventType() string     { return EventTypeAssertionUpdated }
func (e AssertionUpdated) AggregateID() string   { return e.AssertionID }
func (e AssertionUpdated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeAssertionDeleted = "assertion.deleted.v1"

type AssertionDeleted struct {
	event.PerformedBy
	ID            string    `json:"eventId"`
	AssertionID   string    `json:"assertionId"`
	WorkflowID    string    `json:"workflowId"`
	StepID        string    `json:"stepId"`
	ProjectID     string    `json:"projectId"`
	Source        string    `json:"source"`
	Operator      string    `json:"operator"`
	ExpectedValue string    `json:"expectedValue,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e AssertionDeleted) EventID() string       { return e.ID }
func (e AssertionDeleted) EventType() string     { return EventTypeAssertionDeleted }
func (e AssertionDeleted) AggregateID() string   { return e.AssertionID }
func (e AssertionDeleted) OccurredAt() time.Time { return e.Timestamp }
