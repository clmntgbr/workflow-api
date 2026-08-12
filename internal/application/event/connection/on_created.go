package connection

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
)

type ConnectionCreatedHandler struct{}

func NewConnectionCreatedHandler() *ConnectionCreatedHandler {
	return &ConnectionCreatedHandler{}
}

func (h *ConnectionCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainconnection.ConnectionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s connectionId=%s workflowId=%s",
		domainconnection.EventTypeConnectionCreated,
		evt.ID,
		evt.ConnectionID,
		evt.WorkflowID,
	)
	return nil
}
