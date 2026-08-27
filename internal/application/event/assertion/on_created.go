package assertion

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainassertion "go-api/internal/domain/assertion"
)

type AssertionCreatedHandler struct{}

func NewAssertionCreatedHandler() *AssertionCreatedHandler {
	return &AssertionCreatedHandler{}
}

func (h *AssertionCreatedHandler) Handle(_ context.Context, payload []byte) error {
	var evt domainassertion.AssertionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s assertionId=%s stepId=%s",
		domainassertion.EventTypeAssertionCreated,
		evt.ID,
		evt.AssertionID,
		evt.StepID,
	)
	return nil
}
