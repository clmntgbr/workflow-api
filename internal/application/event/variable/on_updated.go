package variable

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainvariable "go-api/internal/domain/variable"
)

type VariableUpdatedHandler struct{}

func NewVariableUpdatedHandler() *VariableUpdatedHandler {
	return &VariableUpdatedHandler{}
}

func (h *VariableUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainvariable.VariableUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s variableId=%s workflowId=%s key=%s",
		domainvariable.EventTypeVariableUpdated,
		evt.ID,
		evt.VariableID,
		evt.WorkflowID,
		evt.Key,
	)
	return nil
}
