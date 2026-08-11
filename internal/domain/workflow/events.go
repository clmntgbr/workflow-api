package workflow

import "time"

const (
	EventTypeWorkflowCreated = "workflow.created.v1"
	EventTypeWorkflowUpdated = "workflow.updated.v1"
	EventTypeWorkflowDeleted = "workflow.deleted.v1"
)

type WorkflowCreated struct {
	ID                      string    `json:"eventId"`
	WorkflowID              string    `json:"workflowId"`
	OrganizationID          string    `json:"organizationId"`
	Name                    string    `json:"name"`
	Description             string    `json:"description"`
	Status                  string    `json:"status"`
	ScheduleIntervalMinutes int       `json:"scheduleIntervalMinutes"`
	Concurrency             int       `json:"concurrency"`
	NotificationsEnabled    bool      `json:"notificationsEnabled"`
	NotifyOnSuccess         bool      `json:"notifyOnSuccess"`
	NotifyOnFailure         bool      `json:"notifyOnFailure"`
	NotifyOnCancel          bool      `json:"notifyOnCancel"`
	Timestamp               time.Time `json:"timestamp"`
}

func (e WorkflowCreated) EventID() string       { return e.ID }
func (e WorkflowCreated) EventType() string     { return EventTypeWorkflowCreated }
func (e WorkflowCreated) AggregateID() string   { return e.WorkflowID }
func (e WorkflowCreated) OccurredAt() time.Time { return e.Timestamp }

type WorkflowUpdated struct {
	ID                      string    `json:"eventId"`
	WorkflowID              string    `json:"workflowId"`
	OrganizationID          string    `json:"organizationId"`
	Name                    string    `json:"name"`
	Description             string    `json:"description"`
	Status                  string    `json:"status"`
	ScheduleIntervalMinutes int       `json:"scheduleIntervalMinutes"`
	Concurrency             int       `json:"concurrency"`
	NotificationsEnabled    bool      `json:"notificationsEnabled"`
	NotifyOnSuccess         bool      `json:"notifyOnSuccess"`
	NotifyOnFailure         bool      `json:"notifyOnFailure"`
	NotifyOnCancel          bool      `json:"notifyOnCancel"`
	Timestamp               time.Time `json:"timestamp"`
}

func (e WorkflowUpdated) EventID() string       { return e.ID }
func (e WorkflowUpdated) EventType() string     { return EventTypeWorkflowUpdated }
func (e WorkflowUpdated) AggregateID() string   { return e.WorkflowID }
func (e WorkflowUpdated) OccurredAt() time.Time { return e.Timestamp }

type WorkflowDeleted struct {
	ID             string    `json:"eventId"`
	WorkflowID     string    `json:"workflowId"`
	OrganizationID string    `json:"organizationId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e WorkflowDeleted) EventID() string       { return e.ID }
func (e WorkflowDeleted) EventType() string     { return EventTypeWorkflowDeleted }
func (e WorkflowDeleted) AggregateID() string   { return e.WorkflowID }
func (e WorkflowDeleted) OccurredAt() time.Time { return e.Timestamp }
