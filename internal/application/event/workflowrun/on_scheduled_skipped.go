package workflowrun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainworkflowrun "go-api/internal/domain/workflowrun"
)

type ScheduledSkippedHandler struct{}

func NewScheduledSkippedHandler() *ScheduledSkippedHandler {
	return &ScheduledSkippedHandler{}
}

func (h *ScheduledSkippedHandler) Handle(ctx context.Context, payload []byte) error {
	_ = ctx
	var evt domainworkflowrun.WorkflowRunScheduledSkipped
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	log.Printf(
		"scheduled workflow run skipped workflowId=%s reason=%s eventId=%s (notification not implemented yet)",
		evt.WorkflowID,
		evt.Reason,
		evt.ID,
	)
	return nil
}
