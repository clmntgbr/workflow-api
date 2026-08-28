package assertion

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime    port.RealtimePublisher
	projectRepo domainproject.ProjectReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	projectRepo domainproject.ProjectReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime:    realtimePublisher,
		projectRepo: projectRepo,
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainassertion.AssertionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainassertion.AssertionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) publishToProjectMembers(
	ctx context.Context,
	projectIDRaw string,
	action string,
	payload any,
) error {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	project, err := h.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if project == nil {
		return messaging.NonRetryable(errProjectNotFound)
	}

	eventType := realtime.EventType(realtime.EntityAssertion, action)
	for _, memberID := range project.MemberIDs {
		if err := h.realtime.PublishToUser(ctx, memberID, eventType, payload); err != nil {
			log.Printf(
				"centrifugo publish failed type=%s projectId=%s userId=%s: %v",
				eventType,
				projectIDRaw,
				memberID.String(),
				err,
			)
			return messaging.Retryable(err)
		}
	}
	return nil
}

type projectNotFoundError struct{}

func (projectNotFoundError) Error() string {
	return "project not found for assertion realtime publish"
}

var errProjectNotFound error = projectNotFoundError{}
