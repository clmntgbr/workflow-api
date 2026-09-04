package project

import (
	"context"
	"encoding/json"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
}

// The project handler never resolves a project from the repository: every event
// it consumes already carries its recipients.
func NewPublishRealtimeHandler(realtimePublisher port.RealtimePublisher) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher, nil),
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityProject, realtime.ActionCreated, evt.CreatedByUserID, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUsers(ctx, realtime.EntityProject, realtime.ActionUpdated, evt.MemberIDs, evt)
}

func (h *PublishRealtimeHandler) OnDeleted(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUsers(ctx, realtime.EntityProject, realtime.ActionDeleted, evt.MemberIDs, evt)
}

func (h *PublishRealtimeHandler) OnMemberAdded(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberAdded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityProject, realtime.ActionMemberAdded, evt.UserID, evt)
}

func (h *PublishRealtimeHandler) OnMemberRemoved(ctx context.Context, payload []byte) error {
	var evt domainproject.ProjectMemberRemoved
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToUser(ctx, realtime.EntityProject, realtime.ActionMemberRemoved, evt.UserID, evt)
}
