package workflow

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowDeletedHandler struct{}

func NewWorkflowDeletedHandler() *WorkflowDeletedHandler {
	return &WorkflowDeletedHandler{}
}

func (h *WorkflowDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainworkflow.WorkflowDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s workflowId=%s organizationId=%s",
		domainworkflow.EventTypeWorkflowDeleted,
		evt.ID,
		evt.WorkflowID,
		evt.OrganizationID,
	)
	return nil
}
