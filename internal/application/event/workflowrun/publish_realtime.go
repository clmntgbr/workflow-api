package workflowrun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	realtime     port.RealtimePublisher
	workflowRepo domainworkflow.WorkflowReadRepository
	projectRepo  domainproject.ProjectReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	workflowRepo domainworkflow.WorkflowReadRepository,
	projectRepo domainproject.ProjectReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		realtime:     realtimePublisher,
		workflowRepo: workflowRepo,
		projectRepo:  projectRepo,
	}
}

func (h *PublishRealtimeHandler) OnStarted(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForWorkflow(ctx, evt.WorkflowID, realtime.ActionStarted, evt)
}

func (h *PublishRealtimeHandler) OnSucceeded(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForWorkflow(ctx, evt.WorkflowID, realtime.ActionSucceeded, evt)
}

func (h *PublishRealtimeHandler) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForWorkflow(ctx, evt.WorkflowID, realtime.ActionFailed, evt)
}

func (h *PublishRealtimeHandler) OnCancelled(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunCancelled
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForWorkflow(ctx, evt.WorkflowID, realtime.ActionCancelled, evt)
}

func (h *PublishRealtimeHandler) publishForWorkflow(
	ctx context.Context,
	workflowIDRaw string,
	action string,
	payload any,
) error {
	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	workflow, err := h.workflowRepo.FindByID(ctx, workflowID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if workflow == nil {
		return messaging.NonRetryable(errWorkflowNotFound)
	}

	org, err := h.projectRepo.FindByID(ctx, workflow.ProjectID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if org == nil {
		return messaging.NonRetryable(errProjectNotFound)
	}

	eventType := realtime.EventType(realtime.EntityWorkflowRun, action)
	for _, memberID := range org.MemberIDs {
		if err := h.realtime.PublishToUser(ctx, memberID, eventType, payload); err != nil {
			log.Printf(
				"centrifugo publish failed type=%s workflowId=%s userId=%s: %v",
				eventType,
				workflowIDRaw,
				memberID.String(),
				err,
			)
			return messaging.Retryable(err)
		}
	}
	return nil
}

type workflowNotFoundError struct{}

func (workflowNotFoundError) Error() string {
	return "workflow not found for workflow run realtime publish"
}

var errWorkflowNotFound error = workflowNotFoundError{}

type projectNotFoundError struct{}

func (projectNotFoundError) Error() string {
	return "project not found for workflow run realtime publish"
}

var errProjectNotFound error = projectNotFoundError{}
