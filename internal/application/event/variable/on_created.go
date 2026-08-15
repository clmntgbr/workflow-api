package variable

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainvariable "go-api/internal/domain/variable"
)

type VariableCreatedHandler struct{}

func NewVariableCreatedHandler() *VariableCreatedHandler {
	return &VariableCreatedHandler{}
}

func (h *VariableCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainvariable.VariableCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s variableId=%s workflowId=%s key=%s",
		domainvariable.EventTypeVariableCreated,
		evt.ID,
		evt.VariableID,
		evt.WorkflowID,
		evt.Key,
	)
	return nil
}
