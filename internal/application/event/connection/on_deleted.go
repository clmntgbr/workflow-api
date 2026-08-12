package connection

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
)

type ConnectionDeletedHandler struct{}

func NewConnectionDeletedHandler() *ConnectionDeletedHandler {
	return &ConnectionDeletedHandler{}
}

func (h *ConnectionDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainconnection.ConnectionDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s connectionId=%s workflowId=%s",
		domainconnection.EventTypeConnectionDeleted,
		evt.ID,
		evt.ConnectionID,
		evt.WorkflowID,
	)
	return nil
}
