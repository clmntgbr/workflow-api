package workflow

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainworkflow "go-api/internal/domain/workflow"
)

type WorkflowCreatedHandler struct{}

func NewWorkflowCreatedHandler() *WorkflowCreatedHandler {
	return &WorkflowCreatedHandler{}
}

func (h *WorkflowCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainworkflow.WorkflowCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s workflowId=%s projectId=%s name=%s",
		domainworkflow.EventTypeWorkflowCreated,
		evt.ID,
		evt.WorkflowID,
		evt.ProjectID,
		evt.Name,
	)
	return nil
}
