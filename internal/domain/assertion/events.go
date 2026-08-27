package assertion

import "time"

const EventTypeAssertionCreated = "assertion.created.v1"

type AssertionCreated struct {
	ID            string    `json:"eventId"`
	AssertionID   string    `json:"assertionId"`
	WorkflowID    string    `json:"workflowId"`
	StepID        string    `json:"stepId"`
	ProjectID     string    `json:"projectId"`
	Name          string    `json:"name"`
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
	ID            string    `json:"eventId"`
	AssertionID   string    `json:"assertionId"`
	WorkflowID    string    `json:"workflowId"`
	StepID        string    `json:"stepId"`
	ProjectID     string    `json:"projectId"`
	Name          string    `json:"name"`
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
