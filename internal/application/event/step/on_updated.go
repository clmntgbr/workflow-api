package step

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainstep "go-api/internal/domain/step"
)

type StepUpdatedHandler struct{}

func NewStepUpdatedHandler() *StepUpdatedHandler {
	return &StepUpdatedHandler{}
}

func (h *StepUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainstep.StepUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s stepId=%s workflowId=%s",
		domainstep.EventTypeStepUpdated,
		evt.ID,
		evt.StepID,
		evt.WorkflowID,
	)
	return nil
}
