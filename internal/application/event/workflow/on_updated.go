package workflow

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowUpdatedHandler struct{}

func NewWorkflowUpdatedHandler() *WorkflowUpdatedHandler {
	return &WorkflowUpdatedHandler{}
}

func (h *WorkflowUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainworkflow.WorkflowUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s workflowId=%s status=%s",
		domainworkflow.EventTypeWorkflowUpdated,
		evt.ID,
		evt.WorkflowID,
		evt.Status,
	)
	return nil
}
