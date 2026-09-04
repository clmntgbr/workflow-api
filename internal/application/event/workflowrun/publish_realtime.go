package workflowrun

import (
	"context"
	"encoding/json"

	"go-api/internal/application/messaging"
	"go-api/internal/application/realtime"
	"go-api/internal/domain/port"
	domainproject "go-api/internal/domain/project"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type PublishRealtimeHandler struct {
	publisher    *realtime.Publisher
	workflowRepo domainworkflow.WorkflowReadRepository
}

func NewPublishRealtimeHandler(
	realtimePublisher port.RealtimePublisher,
	workflowRepo domainworkflow.WorkflowReadRepository,
	projectRepo domainproject.ProjectReadRepository,
) *PublishRealtimeHandler {
	return &PublishRealtimeHandler{
		publisher:    realtime.NewPublisher(realtimePublisher, projectRepo),
		workflowRepo: workflowRepo,
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

func (h *PublishRealtimeHandler) OnFinished(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunFinished
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}
	return h.publishForWorkflow(ctx, evt.WorkflowID, realtime.ActionFinished, evt)
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

	return h.publisher.ToProjectMembers(
		ctx,
		realtime.EntityWorkflowRun,
		action,
		workflow.ProjectID.String(),
		payload,
	)
}

type workflowNotFoundError struct{}

func (workflowNotFoundError) Error() string {
	return "workflow not found for workflow run realtime publish"
}

var errWorkflowNotFound error = workflowNotFoundError{}
