package connection

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Connection struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
	SourceStepID   uuid.UUID
	TargetStepID   uuid.UUID

	events []event.DomainEvent
}

type NewConnectionParams struct {
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
	SourceStepID   uuid.UUID
	TargetStepID   uuid.UUID
}

func NewConnection(p NewConnectionParams) *Connection {
	now := time.Now().UTC()
	c := &Connection{
		ID:             uuid.New(),
		WorkflowID:     p.WorkflowID,
		ProjectID: p.ProjectID,
		SourceStepID:   p.SourceStepID,
		TargetStepID:   p.TargetStepID,
	}
	c.recordEvent(ConnectionCreated{
		ID:             uuid.New().String(),
		ConnectionID:   c.ID.String(),
		WorkflowID:     c.WorkflowID.String(),
		ProjectID: c.ProjectID.String(),
		SourceStepID:   c.SourceStepID.String(),
		TargetStepID:   c.TargetStepID.String(),
		Timestamp:      now,
	})
	return c
}

func (c *Connection) RecordDeletedEvent() {
	c.recordEvent(ConnectionDeleted{
		ID:             uuid.New().String(),
		ConnectionID:   c.ID.String(),
		WorkflowID:     c.WorkflowID.String(),
		ProjectID: c.ProjectID.String(),
		Timestamp:      time.Now().UTC(),
	})
}

func (c *Connection) PullEvents() []event.DomainEvent {
	events := c.events
	c.events = nil
	return events
}

func (c *Connection) recordEvent(evt event.DomainEvent) {
	c.events = append(c.events, evt)
}
