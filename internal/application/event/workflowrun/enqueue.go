package workflowrun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type ExecutorPriorityResolver interface {
	ExecutorPriorityForProject(ctx context.Context, projectID uuid.UUID) (uint8, error)
}

type EnqueueStepRunHandler struct {
	executor         port.StepRunExecutor
	priorityResolver ExecutorPriorityResolver
}

func NewEnqueueStepRunHandler(
	executor port.StepRunExecutor,
	priorityResolver ExecutorPriorityResolver,
) *EnqueueStepRunHandler {
	return &EnqueueStepRunHandler{
		executor:         executor,
		priorityResolver: priorityResolver,
	}
}

func (h *EnqueueStepRunHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunQueued
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	stepRunID, err := uuid.Parse(evt.StepRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	stepID, err := uuid.Parse(evt.StepID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	priority := h.resolvePriority(ctx, evt.ProjectID)

	if err := h.executor.Enqueue(ctx, port.StepRunExecuteJob{
		StepRunID:     stepRunID,
		StepID:        stepID,
		WorkflowRunID: workflowRunID,
		Priority:      priority,
	}); err != nil {
		return messaging.Retryable(err)
	}

	log.Printf(
		"step run queued for executor stepRunId=%s stepId=%s workflowRunId=%s priority=%d",
		evt.StepRunID,
		evt.StepID,
		evt.WorkflowRunID,
		priority,
	)
	return nil
}

func (h *EnqueueStepRunHandler) resolvePriority(ctx context.Context, projectIDRaw string) uint8 {
	if h.priorityResolver == nil {
		return 0
	}
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil || projectID == uuid.Nil {
		return 0
	}

	priority, err := h.priorityResolver.ExecutorPriorityForProject(ctx, projectID)
	if err != nil {
		return 0
	}
	return priority
}
