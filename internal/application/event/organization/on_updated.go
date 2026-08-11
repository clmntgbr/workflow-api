package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainorganization "go-api/internal/domain/organization"
)

type OrganizationUpdatedHandler struct{}

func NewOrganizationUpdatedHandler() *OrganizationUpdatedHandler {
	return &OrganizationUpdatedHandler{}
}

func (h *OrganizationUpdatedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s organizationId=%s name=%s",
		domainorganization.EventTypeOrganizationUpdated,
		evt.ID,
		evt.OrganizationID,
		evt.Name,
	)
	return nil
}
