package steprun

import (
	"time"

	domainstep "go-api/internal/domain/step"
)

const (
	EventTypeStepRunQueued    = "stepRun.queued.v1"
	EventTypeStepRunStarted   = "stepRun.started.v1"
	EventTypeStepRunSucceeded = "stepRun.succeeded.v1"
	EventTypeStepRunFailed    = "stepRun.failed.v1"
)

type StepRunQueued struct {
	ID            string    `json:"eventId"`
	StepRunID     string    `json:"stepRunId"`
	WorkflowRunID string    `json:"workflowRunId"`
	StepID        string    `json:"stepId"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e StepRunQueued) EventID() string       { return e.ID }
func (e StepRunQueued) EventType() string     { return EventTypeStepRunQueued }
func (e StepRunQueued) AggregateID() string   { return e.StepRunID }
func (e StepRunQueued) OccurredAt() time.Time { return e.Timestamp }

type StepRunStarted struct {
	ID             string              `json:"eventId"`
	StepRunID      string              `json:"stepRunId"`
	WorkflowRunID  string              `json:"workflowRunId"`
	StepID         string              `json:"stepId"`
	WorkflowID     string              `json:"workflowId"`
	EndpointID     string              `json:"endpointId"`
	OrganizationID string              `json:"organizationId"`
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	URL            string              `json:"url"`
	Method         string              `json:"method"`
	Headers        map[string]string   `json:"headers"`
	Query          map[string]string   `json:"query"`
	Body           map[string]any      `json:"body"`
	Timeout        int                 `json:"timeout"`
	RetryOnFailure bool                `json:"retryOnFailure"`
	RetryCount     int                 `json:"retryCount"`
	RetryDelay     int                 `json:"retryDelay"`
	Index          string              `json:"index"`
	ExecutionOrder int                 `json:"executionOrder"`
	TreeIndex      int                 `json:"treeIndex"`
	Position       domainstep.Position `json:"position"`
	Status         string              `json:"status"`
	Attempt        int                 `json:"attempt"`
	Timestamp      time.Time           `json:"timestamp"`
}

func (e StepRunStarted) EventID() string       { return e.ID }
func (e StepRunStarted) EventType() string     { return EventTypeStepRunStarted }
func (e StepRunStarted) AggregateID() string   { return e.StepRunID }
func (e StepRunStarted) OccurredAt() time.Time { return e.Timestamp }

type StepRunSucceeded struct {
	ID                 string            `json:"eventId"`
	StepRunID          string            `json:"stepRunId"`
	WorkflowRunID      string            `json:"workflowRunId"`
	StepID             string            `json:"stepId"`
	OrganizationID     string            `json:"organizationId"`
	Status             string            `json:"status"`
	Attempt            int               `json:"attempt"`
	ResponseSnapshot   *ResponseSnapshot `json:"responseSnapshot"`
	ExtractedVariables map[string]any    `json:"extractedVariables"`
	Timestamp          time.Time         `json:"timestamp"`
}

func (e StepRunSucceeded) EventID() string       { return e.ID }
func (e StepRunSucceeded) EventType() string     { return EventTypeStepRunSucceeded }
func (e StepRunSucceeded) AggregateID() string   { return e.StepRunID }
func (e StepRunSucceeded) OccurredAt() time.Time { return e.Timestamp }

type StepRunFailed struct {
	ID               string            `json:"eventId"`
	StepRunID        string            `json:"stepRunId"`
	WorkflowRunID    string            `json:"workflowRunId"`
	StepID           string            `json:"stepId"`
	OrganizationID   string            `json:"organizationId"`
	Status           string            `json:"status"`
	Attempt          int               `json:"attempt"`
	ResponseSnapshot *ResponseSnapshot `json:"responseSnapshot"`
	Error            string            `json:"error"`
	Timestamp        time.Time         `json:"timestamp"`
}

func (e StepRunFailed) EventID() string       { return e.ID }
func (e StepRunFailed) EventType() string     { return EventTypeStepRunFailed }
func (e StepRunFailed) AggregateID() string   { return e.StepRunID }
func (e StepRunFailed) OccurredAt() time.Time { return e.Timestamp }
