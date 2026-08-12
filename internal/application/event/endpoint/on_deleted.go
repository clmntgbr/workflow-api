package endpoint

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointDeletedHandler struct{}

func NewEndpointDeletedHandler() *EndpointDeletedHandler {
	return &EndpointDeletedHandler{}
}

func (h *EndpointDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s endpointId=%s",
		domainendpoint.EventTypeEndpointDeleted,
		evt.ID,
		evt.EndpointID,
	)
	return nil
}
