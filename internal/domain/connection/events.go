package connection

import (
	"time"

	"go-api/internal/domain/event"
)

const EventTypeConnectionCreated = "connection.created.v1"

type ConnectionCreated struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	ConnectionID   string    `json:"connectionId"`
	WorkflowID     string    `json:"workflowId"`
	ProjectID string    `json:"projectId"`
	SourceStepID   string    `json:"sourceStepId"`
	TargetStepID   string    `json:"targetStepId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ConnectionCreated) EventID() string       { return e.ID }
func (e ConnectionCreated) EventType() string     { return EventTypeConnectionCreated }
func (e ConnectionCreated) AggregateID() string   { return e.ConnectionID }
func (e ConnectionCreated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeConnectionDeleted = "connection.deleted.v1"

type ConnectionDeleted struct {
	event.PerformedBy
	ID             string    `json:"eventId"`
	ConnectionID   string    `json:"connectionId"`
	WorkflowID     string    `json:"workflowId"`
	ProjectID string    `json:"projectId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ConnectionDeleted) EventID() string       { return e.ID }
func (e ConnectionDeleted) EventType() string     { return EventTypeConnectionDeleted }
func (e ConnectionDeleted) AggregateID() string   { return e.ConnectionID }
func (e ConnectionDeleted) OccurredAt() time.Time { return e.Timestamp }
