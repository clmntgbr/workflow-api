package connection

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainconnection "go-api/internal/domain/connection"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
	projectRepo domainproject.ProjectReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	projectRepo domainproject.ProjectReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime: realtimePublisher,
		projectRepo: projectRepo,
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainconnection.ConnectionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) OnDeleted(ctx context.Context, payload []byte) error {
	var evt domainconnection.ConnectionDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionDeleted, evt)
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

	org, err := h.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if org == nil {
		return messaging.NonRetryable(errProjectNotFound)
	}

	eventType := realtime.EventType(realtime.EntityConnection, action)
	for _, memberID := range org.MemberIDs {
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
		log.Printf(
			"centrifugo published type=%s projectId=%s userId=%s",
			eventType,
			projectIDRaw,
			memberID.String(),
		)
	}
	return nil
}

type projectNotFoundError struct{}

func (projectNotFoundError) Error() string {
	return "project not found for connection realtime publish"
}

var errProjectNotFound error = projectNotFoundError{}
