package user

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainuser "go-api/internal/domain/user"
)

type UserActiveOrganizationChangedHandler struct{}

func NewUserActiveOrganizationChangedHandler() *UserActiveOrganizationChangedHandler {
	return &UserActiveOrganizationChangedHandler{}
}

func (h *UserActiveOrganizationChangedHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainuser.UserActiveOrganizationChanged
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	log.Printf(
		"event handled %s eventId=%s userId=%s organizationId=%s",
		domainuser.EventTypeUserActiveOrganizationChanged,
		evt.ID,
		evt.UserID,
		evt.OrganizationID,
	)
	return nil
}
