package workflowrun

import (
	"context"
	"errors"
	"fmt"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type Orchestrator struct {
	runRepo      domainworkflowrun.WorkflowRunWriteRepository
	stepRunRepo  domainsteprun.StepRunWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
	variableRepo domainvariable.VariableWriteRepository
	variableRead domainvariable.VariableReadRepository
	outbox       port.OutboxRepository
}

func NewOrchestrator(
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	stepRunRepo domainsteprun.StepRunWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	variableRepo domainvariable.VariableWriteRepository,
	variableRead domainvariable.VariableReadRepository,
	outbox port.OutboxRepository,
) *Orchestrator {
	return &Orchestrator{
		runRepo:      runRepo,
		stepRunRepo:  stepRunRepo,
		stepReadRepo: stepReadRepo,
		connReadRepo: connReadRepo,
		variableRepo: variableRepo,
		variableRead: variableRead,
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

func (h *Orchestrator) buildStepRun(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	step domainstep.StepView,
) (*domainsteprun.StepRun, error) {
	variables, err := h.variableRead.FindByStepID(ctx, step.ID)
	if err != nil {
		return nil, err
	}
	extracts := make([]domainsteprun.VariableExtract, 0, len(variables))
	for _, variable := range variables {
		extracts = append(extracts, domainsteprun.VariableExtract{
			VariableID: variable.ID,
			Key:        variable.Key,
			Path:       variable.Path,
		})
	}

	resolvedURL, resolvedHeaders, resolvedQuery, resolvedBody, resolveErr := domainvariable.ResolveTemplates(
		step.URL,
		step.Headers,
		step.Query,
		step.Body,
		run.Context,
	)

	stepRun := domainsteprun.NewStepRun(domainsteprun.NewStepRunParams{
		WorkflowRunID:    run.ID,
		StepID:           step.ID,
		WorkflowID:       step.WorkflowID,
		EndpointID:       step.EndpointID,
		OrganizationID:   step.OrganizationID,
		Name:             step.Name,
		Description:      step.Description,
		URL:              step.URL,
		Method:           step.Method,
		Headers:          step.Headers,
		Query:            step.Query,
		Body:             step.Body,
		Timeout:          step.Timeout,
		RetryOnFailure:   step.RetryOnFailure,
		RetryCount:       step.RetryCount,
		RetryDelay:       step.RetryDelay,
		Index:            step.Index,
		ExecutionOrder:   step.ExecutionOrder,
		TreeIndex:        step.TreeIndex,
		Position:         step.Position,
		VariableExtracts: extracts,
	})

	var missing domainvariable.MissingVariableError
	if resolveErr != nil {
		if errors.As(resolveErr, &missing) {
			_ = stepRun.MarkFailed(fmt.Sprintf("variable %s not found", missing.Key), nil)
			return stepRun, nil
		}
		return nil, resolveErr
	}

	stepRun.URL = resolvedURL
	stepRun.Headers = resolvedHeaders
	stepRun.Query = resolvedQuery
	stepRun.Body = resolvedBody
	stepRun.Queue()
	return stepRun, nil
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
