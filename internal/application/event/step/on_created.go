package step

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainstep "go-api/internal/domain/step"
)

type StepCreatedHandler struct{}

func NewStepCreatedHandler() *StepCreatedHandler {
	return &StepCreatedHandler{}
}

func (h *StepCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainstep.StepCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s stepId=%s workflowId=%s endpointId=%s",
		domainstep.EventTypeStepCreated,
		evt.ID,
		evt.StepID,
		evt.WorkflowID,
		evt.EndpointID,
	)
	return nil
}
