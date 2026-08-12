package endpoint

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointCreatedHandler struct{}

func NewEndpointCreatedHandler() *EndpointCreatedHandler {
	return &EndpointCreatedHandler{}
}

func (h *EndpointCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s endpointId=%s organizationId=%s name=%s",
		domainendpoint.EventTypeEndpointCreated,
		evt.ID,
		evt.EndpointID,
		evt.OrganizationID,
		evt.Name,
	)
	return nil
}
