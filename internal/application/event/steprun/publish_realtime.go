package steprun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
	domainsteprun "go-api/internal/domain/steprun"

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

func (h *PublishRealtimeHandler) OnStarted(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionStarted, evt)
}

func (h *PublishRealtimeHandler) OnSucceeded(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionSucceeded, evt)
}

func (h *PublishRealtimeHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishToProjectMembers(ctx, evt.ProjectID, realtime.ActionFailed, evt)
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

	eventType := realtime.EventType(realtime.EntityStepRun, action)
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
	}
	return nil
}

type projectNotFoundError struct{}

func (projectNotFoundError) Error() string {
	return "project not found for step run realtime publish"
}

var errProjectNotFound error = projectNotFoundError{}
