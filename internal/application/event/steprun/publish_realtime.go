package steprun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime    port.RealtimePublisher
	orgRepo     domainorganization.OrganizationReadRepository
	stepRunRepo domainsteprun.StepRunReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	orgRepo domainorganization.OrganizationReadRepository,
	stepRunRepo domainsteprun.StepRunReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime:    realtimePublisher,
		orgRepo:     orgRepo,
		stepRunRepo: stepRunRepo,
	}
}

func (h *PublishRealtimeHandler) OnStarted(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToOrganizationMembers(ctx, evt.OrganizationID, realtime.ActionStarted, evt)
}

func (h *PublishRealtimeHandler) OnSucceeded(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	if stepRunID, err := uuid.Parse(evt.StepRunID); err == nil {
		if view, err := h.stepRunRepo.FindByID(ctx, stepRunID); err == nil && view != nil {
			evt.ExtractedVariables = maskExtractedForRealtime(view.ExtractedVariables, view.VariableExtracts)
		}
	}
	return h.publishToOrganizationMembers(ctx, evt.OrganizationID, realtime.ActionSucceeded, evt)
}

func (h *PublishRealtimeHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToOrganizationMembers(ctx, evt.OrganizationID, realtime.ActionFailed, evt)
}

func maskExtractedForRealtime(
	values map[string]any,
	extracts []domainsteprun.VariableExtract,
) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
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

	eventType := realtime.EventType(realtime.EntityStepRun, action)
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
	}
	return nil
}

type organizationNotFoundError struct{}

func (organizationNotFoundError) Error() string {
	return "organization not found for step run realtime publish"
}

var errOrganizationNotFound error = organizationNotFoundError{}
