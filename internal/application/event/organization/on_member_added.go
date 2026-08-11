package organization

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainorganization "go-api/internal/domain/organization"
)

type OrganizationMemberAddedHandler struct{}

func NewOrganizationMemberAddedHandler() *OrganizationMemberAddedHandler {
	return &OrganizationMemberAddedHandler{}
}

func (h *OrganizationMemberAddedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainorganization.OrganizationMemberAdded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s organizationId=%s userId=%s",
		domainorganization.EventTypeOrganizationMemberAdded,
		evt.ID,
		evt.OrganizationID,
		evt.UserID,
	)
	return nil
}
