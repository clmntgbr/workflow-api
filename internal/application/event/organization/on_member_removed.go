package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainorganization "go-api/internal/domain/organization"
)

type OrganizationMemberRemovedHandler struct{}

func NewOrganizationMemberRemovedHandler() *OrganizationMemberRemovedHandler {
	return &OrganizationMemberRemovedHandler{}
}

func (h *OrganizationMemberRemovedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationMemberRemoved
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s organizationId=%s userId=%s",
		domainorganization.EventTypeOrganizationMemberRemoved,
		evt.ID,
		evt.OrganizationID,
		evt.UserID,
	)
	return nil
}
