package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainorganization "go-api/internal/domain/organization"
)

type OrganizationDeletedHandler struct{}

func NewOrganizationDeletedHandler() *OrganizationDeletedHandler {
	return &OrganizationDeletedHandler{}
}

func (h *OrganizationDeletedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s organizationId=%s",
		domainorganization.EventTypeOrganizationDeleted,
		evt.ID,
		evt.OrganizationID,
	)
	return nil
}
