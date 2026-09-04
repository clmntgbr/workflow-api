package endpoint

import (
	"context"
	"encoding/json"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
)

type PublishRealtimeHandler struct {
	publisher *realtime.Publisher
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	projectRepo domainproject.ProjectReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher: realtime.NewPublisher(realtimePublisher, projectRepo),
	}
}

func (h *PublishRealtimeHandler) OnCreated(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToProjectMembers(ctx, realtime.EntityEndpoint, realtime.ActionCreated, evt.ProjectID, evt)
}

func (h *PublishRealtimeHandler) OnUpdated(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToProjectMembers(ctx, realtime.EntityEndpoint, realtime.ActionUpdated, evt.ProjectID, evt)
}

func (h *PublishRealtimeHandler) OnDeleted(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToProjectMembers(ctx, realtime.EntityEndpoint, realtime.ActionDeleted, evt.ProjectID, evt)
}

func (h *PublishRealtimeHandler) OnImported(ctx context.Context, payload []byte) error {
	var evt domainendpoint.EndpointImported
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publisher.ToProjectMembers(ctx, realtime.EntityEndpoint, realtime.ActionImported, evt.ProjectID, evt)
}
