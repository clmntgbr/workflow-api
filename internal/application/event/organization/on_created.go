package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainorganization "go-api/internal/domain/organization"
)

type OrganizationCreatedHandler struct{}

func NewOrganizationCreatedHandler() *OrganizationCreatedHandler {
	return &OrganizationCreatedHandler{}
}

func (h *OrganizationCreatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s organizationId=%s name=%s",
		domainorganization.EventTypeOrganizationCreated,
		evt.ID,
		evt.OrganizationID,
		evt.Name,
	)
	return nil
}
