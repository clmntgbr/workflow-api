package endpoint

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainendpoint "go-api/internal/domain/endpoint"
	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
	orgRepo  domainorganization.OrganizationReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	orgRepo domainorganization.OrganizationReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime: realtimePublisher,
		orgRepo:  orgRepo,
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToOrganizationMembers(ctx, evt.OrganizationID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) publishToOrganizationMembers(
	ctx context.Context,
	organizationIDRaw string,
	action string,
	payload any,
) error {
	organizationID, err := uuid.Parse(organizationIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	org, err := h.orgRepo.FindByID(ctx, organizationID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if org == nil {
		return messaging.NonRetryable(errOrganizationNotFound)
	}

	eventType := realtime.EventType(realtime.EntityEndpoint, action)
	for _, memberID := range org.MemberIDs {
		if err := h.realtime.PublishToUser(ctx, memberID, eventType, payload); err != nil {
			log.Printf(
				"centrifugo publish failed type=%s organizationId=%s userId=%s: %v",
				eventType,
				organizationIDRaw,
				memberID.String(),
				err,
			)
			return messaging.Retryable(err)
		}
		log.Printf(
			"centrifugo published type=%s organizationId=%s userId=%s",
			eventType,
			organizationIDRaw,
			memberID.String(),
		)
	}
	return nil
}

type organizationNotFoundError struct{}

func (organizationNotFoundError) Error() string {
	return "organization not found for endpoint realtime publish"
}

var errOrganizationNotFound error = organizationNotFoundError{}
