package endpoint

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointImportedHandler struct{}

func NewEndpointImportedHandler() *EndpointImportedHandler {
	return &EndpointImportedHandler{}
}

func (h *EndpointImportedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointImported
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s projectId=%s count=%d",
		domainendpoint.EventTypeEndpointImported,
		evt.ID,
		evt.ProjectID,
		evt.Count,
	)
	return nil
}
