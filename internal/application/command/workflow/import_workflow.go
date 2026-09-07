package workflow

import (
	"context"
	"errors"
	"strings"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/messaging"
	"go-api/internal/application/workflowio"
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/event"
	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ImportWorkflowCommand struct {
	UserID    uuid.UUID
	ProjectID uuid.UUID
	Document  workflowio.Document
}

type ImportWorkflowHandler struct {
	workflows   domainworkflow.WorkflowWriteRepository
	steps       domainstep.StepWriteRepository
	connections domainconnection.ConnectionWriteRepository
	variables   domainvariable.VariableWriteRepository
	assertions  domainassertion.AssertionWriteRepository
	outbox      port.OutboxRepository
	assert      *cmdquota.AssertCreateAllowedHandler
}

func NewImportWorkflowHandler(
	workflows domainworkflow.WorkflowWriteRepository,
	steps domainstep.StepWriteRepository,
	connections domainconnection.ConnectionWriteRepository,
	variables domainvariable.VariableWriteRepository,
	assertions domainassertion.AssertionWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *ImportWorkflowHandler {
	return &ImportWorkflowHandler{
		workflows:   workflows,
		steps:       steps,
		connections: connections,
		variables:   variables,
		assertions:  assertions,
		outbox:      outbox,
		assert:      assert,
	}
}

func (h *ImportWorkflowHandler) Handle(ctx context.Context, cmd ImportWorkflowCommand) (*domainworkflow.Workflow, error) {
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if err := cmd.Document.ValidateShape(); err != nil {
		return nil, err
	}

	if err := h.assert.AssertWorkflowGraphCreate(
		ctx,
		cmd.UserID,
		cmd.ProjectID,
		len(cmd.Document.Steps),
		len(cmd.Document.Variables),
		len(cmd.Document.Assertions),
	); err != nil {
		return nil, err
	}
	if err := h.assert.AssertScheduleInterval(
		ctx,
		cmd.UserID,
		cmd.ProjectID,
		domainworkflow.ScheduleType(cmd.Document.Workflow.ScheduleType),
		cmd.Document.Workflow.ScheduleIntervalValue,
		domainworkflow.ScheduleUnit(cmd.Document.Workflow.ScheduleIntervalUnit),
	); err != nil {
		return nil, err
	}

	for _, step := range cmd.Document.Steps {
		if step.Type != string(domainstep.TypeHTTP) {
			continue
		}
		if err := h.assert.AssertStepHTTPConfig(
			ctx,
			cmd.UserID,
			cmd.ProjectID,
			step.Timeout,
			step.RetryCount,
		); err != nil {
			return nil, err
		}
	}

	workflow, err := domainworkflow.NewWorkflow(domainworkflow.NewWorkflowParams{
		Name:                  strings.TrimSpace(cmd.Document.Workflow.Name),
		Description:           cmd.Document.Workflow.Description,
		ProjectID:             cmd.ProjectID,
		ScheduleType:          domainworkflow.ScheduleType(cmd.Document.Workflow.ScheduleType),
		ScheduleIntervalValue: cmd.Document.Workflow.ScheduleIntervalValue,
		ScheduleIntervalUnit:  domainworkflow.ScheduleUnit(cmd.Document.Workflow.ScheduleIntervalUnit),
		ScheduleAt:            cmd.Document.Workflow.ScheduleAt,
		ScheduleTimezone:      cmd.Document.Workflow.ScheduleTimezone,
		Concurrency:           cmd.Document.Workflow.Concurrency,
		NotificationsEnabled:  cmd.Document.Workflow.NotificationsEnabled,
		NotifyOnSuccess:       cmd.Document.Workflow.NotifyOnSuccess,
		NotifyOnFailure:       cmd.Document.Workflow.NotifyOnFailure,
		NotifyOnCancel:        cmd.Document.Workflow.NotifyOnCancel,
	})
	if err != nil {
		return nil, err
	}

	idByRef := make(map[string]uuid.UUID, len(cmd.Document.Steps))
	positioned := make([]domainstep.PositionedStep, 0, len(cmd.Document.Steps))
	createdAt := time.Now().UTC()
	for _, item := range cmd.Document.Steps {
		ref := strings.TrimSpace(item.Ref)
		id := uuid.New()
		idByRef[ref] = id
		positioned = append(positioned, domainstep.PositionedStep{
			ID: id,
			Position: domainstep.Position{
				X: item.Position.X,
				Y: item.Position.Y,
			},
			CreatedAt: createdAt,
		})
	}

	edges := make([]domainstep.GraphEdge, 0, len(cmd.Document.Connections))
	for _, item := range cmd.Document.Connections {
		sourceID, okSource := idByRef[strings.TrimSpace(item.SourceRef)]
		targetID, okTarget := idByRef[strings.TrimSpace(item.TargetRef)]
		if !okSource || !okTarget {
			return nil, workflowio.Invalid("unknown connection ref")
		}
		edges = append(edges, domainstep.GraphEdge{SourceStepID: sourceID, TargetStepID: targetID})
	}

	ordering := domainstep.CalculateOrderingByPosition(positioned, edges)
	executionOrderByStepID := make(map[uuid.UUID]int, len(ordering))
	for stepID, values := range ordering {
		executionOrderByStepID[stepID] = values.ExecutionOrder
	}
	treeIndices := domainstep.CalculateTreeIndices(executionOrderByStepID, edges)

	stepByRef, stepEntities, err := buildImportedSteps(cmd.Document.Steps, workflow, idByRef, ordering, treeIndices)
	if err != nil {
		return nil, err
	}

	connEntities, connViews, err := buildImportedConnections(cmd.Document.Connections, stepByRef, workflow)
	if err != nil {
		return nil, err
	}

	stepViews := make([]domainstep.StepView, 0, len(stepEntities))
	for _, step := range stepEntities {
		stepViews = append(stepViews, stepViewFromEntity(step))
	}
	if err := domainconnection.ValidateWorkflowConnections(stepViews, connViews); err != nil {
		return nil, err
	}

	variableEntities, err := buildImportedVariables(cmd.Document.Variables, stepByRef, workflow)
	if err != nil {
		return nil, err
	}
	assertionEntities, err := buildImportedAssertions(cmd.Document.Assertions, stepByRef, workflow)
	if err != nil {
		return nil, err
	}

	err = h.workflows.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.workflows.Save(txCtx, workflow); err != nil {
			return err
		}
		for _, step := range stepEntities {
			if err := h.steps.Save(txCtx, step); err != nil {
				return err
			}
		}
		if len(ordering) > 0 {
			if err := h.steps.UpdateOrdering(txCtx, ordering); err != nil {
				return err
			}
			if err := h.steps.UpdateTreeIndices(txCtx, treeIndices); err != nil {
				return err
			}
		}
		for _, conn := range connEntities {
			if err := h.connections.Save(txCtx, conn); err != nil {
				return err
			}
		}
		for _, variable := range variableEntities {
			if err := h.variables.Save(txCtx, variable); err != nil {
				if errors.Is(err, domainvariable.ErrDuplicateKey) {
					return err
				}
				return errors.New("failed to import workflow")
			}
		}
		for _, assertion := range assertionEntities {
			if err := h.assertions.Save(txCtx, assertion); err != nil {
				return err
			}
		}

		var events []event.DomainEvent
		events = append(events, messaging.WithPerformedBy(workflow.PullEvents(), cmd.UserID)...)
		for _, step := range stepEntities {
			events = append(events, messaging.WithPerformedBy(step.PullEvents(), cmd.UserID)...)
		}
		for _, conn := range connEntities {
			events = append(events, messaging.WithPerformedBy(conn.PullEvents(), cmd.UserID)...)
		}
		for _, variable := range variableEntities {
			events = append(events, messaging.WithPerformedBy(variable.PullEvents(), cmd.UserID)...)
		}
		for _, assertion := range assertionEntities {
			events = append(events, messaging.WithPerformedBy(assertion.PullEvents(), cmd.UserID)...)
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		if errors.Is(err, domainvariable.ErrDuplicateKey) {
			return nil, err
		}
		if errors.Is(err, workflowio.ErrInvalidDocument) || errors.Is(err, workflowio.ErrUnsupportedFormat) {
			return nil, err
		}
		return nil, errors.New("failed to import workflow")
	}

	return workflow, nil
}

func buildImportedSteps(
	items []workflowio.Step,
	workflow *domainworkflow.Workflow,
	idByRef map[string]uuid.UUID,
	ordering map[uuid.UUID]domainstep.StepOrdering,
	treeIndices map[uuid.UUID]int,
) (map[string]*domainstep.Step, []*domainstep.Step, error) {
	stepByRef := make(map[string]*domainstep.Step, len(items))
	entities := make([]*domainstep.Step, 0, len(items))
	for _, item := range items {
		ref := strings.TrimSpace(item.Ref)
		id := idByRef[ref]
		values := ordering[id]
		treeIndex := treeIndices[id]
		stepType, err := domainstep.ParseType(item.Type)
		if err != nil {
			return nil, nil, workflowio.Invalid("invalid step type")
		}

		var step *domainstep.Step
		switch stepType {
		case domainstep.TypeHTTP:
			step, err = domainstep.NewHTTPStep(domainstep.NewHTTPStepParams{
				ID:             id,
				WorkflowID:     workflow.ID,
				ProjectID:      workflow.ProjectID,
				Name:           item.Name,
				Description:    item.Description,
				URL:            item.URL,
				Method:         item.Method,
				Headers:        item.Headers,
				Query:          httpquery.Clone(item.Query),
				Body:           item.Body,
				Timeout:        item.Timeout,
				RetryOnFailure: item.RetryOnFailure,
				RetryCount:     item.RetryCount,
				RetryDelay:     item.RetryDelay,
				Index:          values.Index,
				ExecutionOrder: values.ExecutionOrder,
				TreeIndex:      treeIndex,
				Position: domainstep.Position{
					X: item.Position.X,
					Y: item.Position.Y,
				},
			})
		case domainstep.TypeDelay:
			step, err = domainstep.NewDelayStep(domainstep.NewDelayStepParams{
				ID:                   id,
				WorkflowID:           workflow.ID,
				ProjectID:            workflow.ProjectID,
				Name:                 item.Name,
				DelayDurationSeconds: item.DelayDurationSeconds,
				Index:                values.Index,
				ExecutionOrder:       values.ExecutionOrder,
				TreeIndex:            treeIndex,
				Position: domainstep.Position{
					X: item.Position.X,
					Y: item.Position.Y,
				},
			})
		case domainstep.TypeCondition:
			step, err = domainstep.NewConditionStep(domainstep.NewConditionStepParams{
				ID:             id,
				WorkflowID:     workflow.ID,
				ProjectID:      workflow.ProjectID,
				Name:           item.Name,
				Expression:     item.Expression,
				Index:          values.Index,
				ExecutionOrder: values.ExecutionOrder,
				TreeIndex:      treeIndex,
				Position: domainstep.Position{
					X: item.Position.X,
					Y: item.Position.Y,
				},
			})
		}
		if err != nil {
			return nil, nil, err
		}
		stepByRef[ref] = step
		entities = append(entities, step)
	}
	return stepByRef, entities, nil
}

func buildImportedConnections(
	items []workflowio.Connection,
	stepByRef map[string]*domainstep.Step,
	workflow *domainworkflow.Workflow,
) ([]*domainconnection.Connection, []domainconnection.ConnectionView, error) {
	entities := make([]*domainconnection.Connection, 0, len(items))
	views := make([]domainconnection.ConnectionView, 0, len(items))
	for _, item := range items {
		source := stepByRef[strings.TrimSpace(item.SourceRef)]
		target := stepByRef[strings.TrimSpace(item.TargetRef)]
		if source == nil || target == nil {
			return nil, nil, workflowio.Invalid("unknown connection ref")
		}

		var branch *domainconnection.ConditionBranch
		if item.Branch != nil && strings.TrimSpace(*item.Branch) != "" {
			parsed, err := domainconnection.ParseConditionBranch(strings.TrimSpace(*item.Branch))
			if err != nil {
				return nil, nil, err
			}
			branch = &parsed
		}

		conn := domainconnection.NewConnection(domainconnection.NewConnectionParams{
			WorkflowID:   workflow.ID,
			ProjectID:    workflow.ProjectID,
			SourceStepID: source.ID,
			TargetStepID: target.ID,
			Branch:       branch,
		})
		entities = append(entities, conn)
		views = append(views, domainconnection.ConnectionView{
			ID:           conn.ID,
			WorkflowID:   conn.WorkflowID,
			ProjectID:    conn.ProjectID,
			SourceStepID: conn.SourceStepID,
			TargetStepID: conn.TargetStepID,
			Branch:       conn.Branch,
		})
	}
	return entities, views, nil
}

func buildImportedVariables(
	items []workflowio.Variable,
	stepByRef map[string]*domainstep.Step,
	workflow *domainworkflow.Workflow,
) ([]*domainvariable.Variable, error) {
	entities := make([]*domainvariable.Variable, 0, len(items))
	for _, item := range items {
		kind, err := domainvariable.ParseKind(item.Kind)
		if err != nil {
			return nil, workflowio.Invalid("invalid variable kind")
		}
		var stepID *uuid.UUID
		path := strings.TrimSpace(item.Path)
		var value any
		switch kind {
		case domainvariable.KindExtracted:
			step := stepByRef[strings.TrimSpace(item.StepRef)]
			if step == nil {
				return nil, workflowio.Invalid("unknown variable stepRef")
			}
			id := step.ID
			stepID = &id
			value = nil
		case domainvariable.KindStatic:
			stepID = nil
			path = ""
			value = item.Value
		}
		variable, err := domainvariable.NewVariable(domainvariable.NewVariableParams{
			Name:        strings.TrimSpace(item.Name),
			Key:         strings.TrimSpace(item.Key),
			Description: strings.TrimSpace(item.Description),
			Kind:        kind,
			Path:        path,
			Value:       value,
			StepID:      stepID,
			WorkflowID:  workflow.ID,
			ProjectID:   workflow.ProjectID,
		})
		if err != nil {
			return nil, err
		}
		entities = append(entities, variable)
	}
	return entities, nil
}

func buildImportedAssertions(
	items []workflowio.Assertion,
	stepByRef map[string]*domainstep.Step,
	workflow *domainworkflow.Workflow,
) ([]*domainassertion.Assertion, error) {
	entities := make([]*domainassertion.Assertion, 0, len(items))
	for _, item := range items {
		step := stepByRef[strings.TrimSpace(item.StepRef)]
		if step == nil {
			return nil, workflowio.Invalid("unknown assertion stepRef")
		}
		source, err := domainassertion.ParseSource(item.Source)
		if err != nil {
			return nil, err
		}
		operator, err := domainassertion.ParseOperator(item.Operator)
		if err != nil {
			return nil, err
		}
		assertion, err := domainassertion.NewAssertion(domainassertion.NewAssertionParams{
			Description:   item.Description,
			Source:        source,
			Path:          item.Path,
			Operator:      operator,
			ExpectedValue: item.ExpectedValue,
			StepID:        step.ID,
			WorkflowID:    workflow.ID,
			ProjectID:     workflow.ProjectID,
		})
		if err != nil {
			return nil, err
		}
		entities = append(entities, assertion)
	}
	return entities, nil
}

func stepViewFromEntity(step *domainstep.Step) domainstep.StepView {
	return domainstep.StepView{
		ID:                   step.ID,
		WorkflowID:           step.WorkflowID,
		ProjectID:            step.ProjectID,
		Type:                 step.Type,
		DelayDurationSeconds: step.DelayDurationSeconds,
		Expression:           step.Expression,
		Name:                 step.Name,
		Description:          step.Description,
		URL:                  step.URL,
		Method:               step.Method,
		Headers:              step.Headers,
		Query:                step.Query,
		Body:                 step.Body,
		Timeout:              step.Timeout,
		RetryOnFailure:       step.RetryOnFailure,
		RetryCount:           step.RetryCount,
		RetryDelay:           step.RetryDelay,
		Index:                step.Index,
		ExecutionOrder:       step.ExecutionOrder,
		TreeIndex:            step.TreeIndex,
		Position:             step.Position,
		Status:               step.Status,
		CreatedAt:            step.CreatedAt,
		UpdatedAt:            step.UpdatedAt,
	}
}
