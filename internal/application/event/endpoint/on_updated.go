package endpoint

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointUpdatedHandler struct{}

func NewEndpointUpdatedHandler() *EndpointUpdatedHandler {
	return &EndpointUpdatedHandler{}
}

func (h *EndpointUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s endpointId=%s status=%s",
		domainendpoint.EventTypeEndpointUpdated,
		evt.ID,
		evt.EndpointID,
		evt.Status,
	)
	return nil
}
