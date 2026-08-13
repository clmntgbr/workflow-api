package workflowrun

import "time"

const (
	EventTypeWorkflowRunStarted          = "workflowRun.started.v1"
	EventTypeWorkflowRunSucceeded        = "workflowRun.succeeded.v1"
	EventTypeWorkflowRunFailed           = "workflowRun.failed.v1"
	EventTypeWorkflowRunCancelled        = "workflowRun.cancelled.v1"
	EventTypeWorkflowRunScheduledSkipped = "workflowRun.scheduledSkipped.v1"
)

type WorkflowRunStarted struct {
	ID                string    `json:"eventId"`
	WorkflowRunID     string    `json:"workflowRunId"`
	WorkflowID        string    `json:"workflowId"`
	Status            string    `json:"status"`
	TriggeredBy       string    `json:"triggeredBy"`
	TriggeredByUserID *string   `json:"triggeredByUserId"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e WorkflowRunStarted) EventID() string       { return e.ID }
func (e WorkflowRunStarted) EventType() string     { return EventTypeWorkflowRunStarted }
func (e WorkflowRunStarted) AggregateID() string   { return e.WorkflowRunID }
func (e WorkflowRunStarted) OccurredAt() time.Time { return e.Timestamp }

type WorkflowRunSucceeded struct {
	ID            string    `json:"eventId"`
	WorkflowRunID string    `json:"workflowRunId"`
	WorkflowID    string    `json:"workflowId"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e WorkflowRunSucceeded) EventID() string       { return e.ID }
func (e WorkflowRunSucceeded) EventType() string     { return EventTypeWorkflowRunSucceeded }
func (e WorkflowRunSucceeded) AggregateID() string   { return e.WorkflowRunID }
func (e WorkflowRunSucceeded) OccurredAt() time.Time { return e.Timestamp }

type WorkflowRunFailed struct {
	ID            string    `json:"eventId"`
	WorkflowRunID string    `json:"workflowRunId"`
	WorkflowID    string    `json:"workflowId"`
	Status        string    `json:"status"`
	Error         string    `json:"error"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e WorkflowRunFailed) EventID() string       { return e.ID }
func (e WorkflowRunFailed) EventType() string     { return EventTypeWorkflowRunFailed }
func (e WorkflowRunFailed) AggregateID() string   { return e.WorkflowRunID }
func (e WorkflowRunFailed) OccurredAt() time.Time { return e.Timestamp }

type WorkflowRunCancelled struct {
	ID            string    `json:"eventId"`
	WorkflowRunID string    `json:"workflowRunId"`
	WorkflowID    string    `json:"workflowId"`
	Status        string    `json:"status"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e WorkflowRunCancelled) EventID() string       { return e.ID }
func (e WorkflowRunCancelled) EventType() string     { return EventTypeWorkflowRunCancelled }
func (e WorkflowRunCancelled) AggregateID() string   { return e.WorkflowRunID }
func (e WorkflowRunCancelled) OccurredAt() time.Time { return e.Timestamp }

// WorkflowRunScheduledSkipped is emitted when the scheduler would start a run
// but one is already pending/running for that workflow.
type WorkflowRunScheduledSkipped struct {
	ID         string    `json:"eventId"`
	WorkflowID string    `json:"workflowId"`
	Reason     string    `json:"reason"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e WorkflowRunScheduledSkipped) EventID() string       { return e.ID }
func (e WorkflowRunScheduledSkipped) EventType() string     { return EventTypeWorkflowRunScheduledSkipped }
func (e WorkflowRunScheduledSkipped) AggregateID() string   { return e.WorkflowID }
func (e WorkflowRunScheduledSkipped) OccurredAt() time.Time { return e.Timestamp }

const ScheduledSkipReasonAlreadyInProgress = "already_in_progress"
