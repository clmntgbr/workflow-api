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

type EnqueueStepRunHandler struct {
	executor port.StepRunExecutor
}

func NewEnqueueStepRunHandler(executor port.StepRunExecutor) *EnqueueStepRunHandler {
	return &EnqueueStepRunHandler{executor: executor}
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

	if err := h.executor.Enqueue(ctx, port.StepRunExecuteJob{
		StepRunID:     stepRunID,
		StepID:        stepID,
		WorkflowRunID: workflowRunID,
	}); err != nil {
		return messaging.Retryable(err)
	}

	log.Printf(
		"step run queued for executor stepRunId=%s stepId=%s workflowRunId=%s",
		evt.StepRunID,
		evt.StepID,
		evt.WorkflowRunID,
	)
	return nil
}
