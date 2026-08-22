package workflowrun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

func (h *Orchestrator) OnStarted(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	runID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	run, err := h.runRepo.GetByID(ctx, runID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if run == nil {
		return messaging.NonRetryable(errWorkflowRunNotFound)
	}
	if run.Status.IsTerminal() {
		return nil
	}

	steps, connections, err := h.loadGraph(ctx, run.WorkflowID)
	if err != nil {
		return err
	}

	existing, err := h.stepRunRepo.FindByWorkflowRunID(ctx, run.ID)
	if err != nil {
		return messaging.Retryable(err)
	}
	existingByStep := stepRunsByStepID(existing)

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if run.Status == domainworkflowrun.StatusPending {
			if err := run.MarkRunning(); err != nil {
				return err
			}
			if err := h.runRepo.Update(txCtx, run); err != nil {
				return err
			}
		}

		if len(steps) == 0 {
			if !run.Status.IsTerminal() {
				if err := run.MarkSucceeded(); err != nil {
					return err
				}
				if err := h.runRepo.Update(txCtx, run); err != nil {
					return err
				}
			}
			return h.outbox.StoreEvents(txCtx, run.PullEvents())
		}

		created := make([]*domainsteprun.StepRun, 0)
		for _, step := range rootSteps(steps, connections) {
			if _, ok := existingByStep[step.ID]; ok {
				continue
			}
			stepRun, err := h.buildStepRun(txCtx, run, step)
			if err != nil {
				return err
			}
			if err := h.stepRunRepo.Save(txCtx, stepRun); err != nil {
				return err
			}
			created = append(created, stepRun)
		}

		return h.outbox.StoreEvents(txCtx, collectEvents(run, created...))
	})
	if err != nil {
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s workflowRunId=%s workflowId=%s",
		domainworkflowrun.EventTypeWorkflowRunStarted,
		evt.ID,
		evt.WorkflowRunID,
		evt.WorkflowID,
	)
	return nil
}

func (h *Orchestrator) OnSucceeded(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	stepID, err := uuid.Parse(evt.StepID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	stepRunID, err := uuid.Parse(evt.StepRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	return h.advanceAfterStep(ctx, workflowRunID, stepID, stepRunID, evt.ExtractedVariables)
}

func (h *Orchestrator) OnFailed(ctx context.Context, payload []byte) error {
	var evt domainsteprun.StepRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	stepID, err := uuid.Parse(evt.StepID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	return h.advanceAfterStep(ctx, workflowRunID, stepID, uuid.Nil, nil)
}

func (h *Orchestrator) advanceAfterStep(
	ctx context.Context,
	workflowRunID uuid.UUID,
	completedStepID uuid.UUID,
	completedStepRunID uuid.UUID,
	extractedVariables map[string]any,
) error {
	run, err := h.runRepo.GetByID(ctx, workflowRunID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if run == nil {
		return messaging.NonRetryable(errWorkflowRunNotFound)
	}
	if run.Status.IsTerminal() {
		return nil
	}

	steps, connections, err := h.loadGraph(ctx, run.WorkflowID)
	if err != nil {
		return err
	}

	existing, err := h.stepRunRepo.FindByWorkflowRunID(ctx, run.ID)
	if err != nil {
		return messaging.Retryable(err)
	}
	runsByStep := stepRunsByStepID(existing)
	stepsByID := stepByID(steps)

	err = h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if len(extractedVariables) > 0 {
			run.MergeContext(extractedVariables)
			if err := h.runRepo.Update(txCtx, run); err != nil {
				return err
			}
		}

		changed := make([]*domainsteprun.StepRun, 0)

		for _, targetID := range outgoingIDs(completedStepID, connections) {
			if _, exists := runsByStep[targetID]; exists {
				continue
			}
			target, ok := stepsByID[targetID]
			if !ok {
				continue
			}

			switch {
			case canEnqueue(targetID, connections, runsByStep):
				stepRun, err := h.buildStepRun(txCtx, run, target)
				if err != nil {
					return err
				}
				if err := h.stepRunRepo.Save(txCtx, stepRun); err != nil {
					return err
				}
				runsByStep[targetID] = stepRun
				changed = append(changed, stepRun)
			case shouldSkip(targetID, connections, runsByStep):
				skipped, err := h.skipBranch(txCtx, run, target, stepsByID, connections, runsByStep)
				if err != nil {
					return err
				}
				changed = append(changed, skipped...)
			}
		}

		if err := h.finalizeRun(txCtx, run, steps, connections, runsByStep); err != nil {
			return err
		}

		events := collectEvents(run, changed...)
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return messaging.Retryable(err)
	}
	_ = completedStepRunID
	return nil
}


func (h *Orchestrator) skipBranch(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	step domainstep.StepView,
	stepsByID map[uuid.UUID]domainstep.StepView,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) ([]*domainsteprun.StepRun, error) {
	created := make([]*domainsteprun.StepRun, 0)
	queue := []domainstep.StepView{step}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, exists := runsByStep[current.ID]; exists {
			continue
		}
		if !shouldSkip(current.ID, connections, runsByStep) && current.ID != step.ID {
			continue
		}

		stepRun := domainsteprun.NewStepRun(domainsteprun.NewStepRunParams{
			WorkflowRunID:  run.ID,
			StepID:         current.ID,
			WorkflowID:     current.WorkflowID,
			EndpointID:     current.EndpointID,
			ProjectID: current.ProjectID,
			Name:           current.Name,
			Description:    current.Description,
			URL:            current.URL,
			Method:         current.Method,
			Headers:        current.Headers,
			Query:          current.Query,
			Body:           current.Body,
			Timeout:        current.Timeout,
			RetryOnFailure: current.RetryOnFailure,
			RetryCount:     current.RetryCount,
			RetryDelay:     current.RetryDelay,
			Index:          current.Index,
			ExecutionOrder: current.ExecutionOrder,
			TreeIndex:      current.TreeIndex,
			Position:       current.Position,
		})
		if err := stepRun.MarkSkipped(); err != nil {
			return nil, err
		}
		if err := h.stepRunRepo.Save(ctx, stepRun); err != nil {
			return nil, err
		}
		runsByStep[current.ID] = stepRun
		created = append(created, stepRun)

		for _, nextID := range outgoingIDs(current.ID, connections) {
			next, ok := stepsByID[nextID]
			if !ok {
				continue
			}
			if _, exists := runsByStep[nextID]; exists {
				continue
			}
			queue = append(queue, next)
		}
	}

	return created, nil
}

func (h *Orchestrator) finalizeRun(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	steps []domainstep.StepView,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) error {
	reachable := reachableStepIDs(steps, connections)
	if len(reachable) == 0 {
		if run.Status.IsTerminal() {
			return nil
		}
		if err := run.MarkSucceeded(); err != nil {
			return err
		}
		return h.runRepo.Update(ctx, run)
	}

	anyFailed := false
	for stepID := range reachable {
		stepRun, ok := runsByStep[stepID]
		if !ok || !stepRun.Status.IsTerminal() {
			return nil
		}
		if stepRun.Status == domainsteprun.StatusFailed {
			anyFailed = true
		}
	}

	if run.Status.IsTerminal() {
		return nil
	}
	if anyFailed {
		if err := run.MarkFailed("one or more steps failed"); err != nil {
			return err
		}
	} else {
		if err := run.MarkSucceeded(); err != nil {
			return err
		}
	}
	return h.runRepo.Update(ctx, run)
}

func canEnqueue(
	stepID uuid.UUID,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) bool {
	incoming := incomingIDs(stepID, connections)
	if len(incoming) == 0 {
		return true
	}
	for _, parentID := range incoming {
		parent, ok := runsByStep[parentID]
		if !ok || parent.Status != domainsteprun.StatusSuccess {
			return false
		}
	}
	return true
}

func shouldSkip(
	stepID uuid.UUID,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) bool {
	incoming := incomingIDs(stepID, connections)
	if len(incoming) == 0 {
		return false
	}

	blocked := false
	for _, parentID := range incoming {
		parent, ok := runsByStep[parentID]
		if !ok || !parent.Status.IsTerminal() {
			return false
		}
		if parent.Status == domainsteprun.StatusFailed || parent.Status == domainsteprun.StatusSkipped {
			blocked = true
		}
	}
	return blocked
}

type workflowRunNotFoundError struct{}

func (workflowRunNotFoundError) Error() string { return "workflow run not found" }

var errWorkflowRunNotFound error = workflowRunNotFoundError{}
