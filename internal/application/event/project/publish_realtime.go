package project

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime port.RealtimePublisher
}

func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{realtime: realtimePublisher}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	userID, err := uuid.Parse(evt.CreatedByUserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	return h.publishToUser(ctx, userID, realtime.ActionCreated, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	return h.publishToMembers(ctx, evt.MemberIDs, realtime.ActionUpdated, evt)
}

func (h *PublishRealtimeHandler) OnDeleted(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	return h.publishToMembers(ctx, evt.MemberIDs, realtime.ActionDeleted, evt)
}

func (h *PublishRealtimeHandler) OnMemberAdded(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberAdded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	return h.publishToUser(ctx, userID, realtime.ActionMemberAdded, evt)
}

func (h *PublishRealtimeHandler) OnMemberRemoved(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberRemoved
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	userID, err := uuid.Parse(evt.UserID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	return h.publishToUser(ctx, userID, realtime.ActionMemberRemoved, evt)
}

func (h *PublishRealtimeHandler) publishToMembers(
	ctx context.Context,
	memberIDs []string,
	action string,
	payload any,
) error {
	for _, memberIDRaw := range memberIDs {
		userID, err := uuid.Parse(memberIDRaw)
		if err != nil {
			return messaging.NonRetryable(err)
		}
		if err := h.publishToUser(ctx, userID, action, payload); err != nil {
			return err
		}
	}
	return nil
}

func (h *PublishRealtimeHandler) publishToUser(
	ctx context.Context,
	userID uuid.UUID,
	action string,
	payload any,
) error {
	eventType := realtime.EventType(realtime.EntityProject, action)
	if err := h.realtime.PublishToUser(ctx, userID, eventType, payload); err != nil {
		log.Printf("centrifugo publish failed type=%s userId=%s: %v", eventType, userID.String(), err)
		return messaging.Retryable(err)
	}
	log.Printf("centrifugo published type=%s userId=%s", eventType, userID.String())
	return nil
}
