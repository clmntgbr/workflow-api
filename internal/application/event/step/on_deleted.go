package step

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainstep "go-api/internal/domain/step"
)

type StepDeletedHandler struct{}

func NewStepDeletedHandler() *StepDeletedHandler {
	return &StepDeletedHandler{}
}

func (h *StepDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainstep.StepDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s stepId=%s workflowId=%s",
		domainstep.EventTypeStepDeleted,
		evt.ID,
		evt.StepID,
		evt.WorkflowID,
	)
	return nil
}
