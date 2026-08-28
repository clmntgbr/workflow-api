package assertion

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainassertion "go-api/internal/domain/assertion"
)

type AssertionUpdatedHandler struct{}

func NewAssertionUpdatedHandler() *AssertionUpdatedHandler {
	return &AssertionUpdatedHandler{}
}

func (h *AssertionUpdatedHandler) Handle(_ context.Context, payload []byte) error {
	var evt domainassertion.AssertionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s assertionId=%s stepId=%s",
		domainassertion.EventTypeAssertionUpdated,
		evt.ID,
		evt.AssertionID,
		evt.StepID,
	)
	return nil
}
