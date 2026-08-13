package workflowrun

import (
	"context"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type Orchestrator struct {
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	stepRunRepo  domainsteprun.StepRunWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
	outbox       port.OutboxRepository
}

func NewOrchestrator(
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	stepRunRepo domainsteprun.StepRunWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	outbox port.OutboxRepository,
) *Orchestrator {
	return &Orchestrator{
		runRepo:      runRepo,
		stepRunRepo:  stepRunRepo,
		stepReadRepo: stepReadRepo,
		connReadRepo: connReadRepo,
		outbox:       outbox,
	}
}

func (h *Orchestrator) loadGraph(
	ctx context.Context,
	workflowID uuid.UUID,
) ([]domainstep.StepView, []domainconnection.ConnectionView, error) {
	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, nil, messaging.Retryable(err)
	}
	connections, err := h.connReadRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, nil, messaging.Retryable(err)
	}
	return steps, connections, nil
}

func newStepRunFromStep(workflowRunID uuid.UUID, step domainstep.StepView) *domainsteprun.StepRun {
	return domainsteprun.NewStepRun(domainsteprun.NewStepRunParams{
		WorkflowRunID:  workflowRunID,
		StepID:         step.ID,
		WorkflowID:     step.WorkflowID,
		EndpointID:     step.EndpointID,
		OrganizationID: step.OrganizationID,
		Name:           step.Name,
		Description:    step.Description,
		URL:            step.URL,
		Method:         step.Method,
		Headers:        step.Headers,
		Query:          step.Query,
		Body:           step.Body,
		Timeout:        step.Timeout,
		RetryOnFailure: step.RetryOnFailure,
		RetryCount:     step.RetryCount,
		RetryDelay:     step.RetryDelay,
		Index:          step.Index,
		ExecutionOrder: step.ExecutionOrder,
		TreeIndex:      step.TreeIndex,
		Position:       step.Position,
	})
}

func stepRunsByStepID(runs []*domainsteprun.StepRun) map[uuid.UUID]*domainsteprun.StepRun {
	out := make(map[uuid.UUID]*domainsteprun.StepRun, len(runs))
	for _, run := range runs {
		out[run.StepID] = run
	}
	return out
}

func collectEvents(run *domainworkflowrun.WorkflowRun, stepRuns ...*domainsteprun.StepRun) []event.DomainEvent {
	events := run.PullEvents()
	for _, stepRun := range stepRuns {
		if stepRun == nil {
			continue
		}
		events = append(events, stepRun.PullEvents()...)
	}
	return events
}
