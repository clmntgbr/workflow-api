package activitylog

import (
	"encoding/json"
	"fmt"
	"time"

	domainactivitylog "go-api/internal/domain/activitylog"
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

func Project(eventType string, payload []byte) (*domainactivitylog.Entry, error) {
	switch eventType {
	case domainworkflow.EventTypeWorkflowCreated:
		return projectWorkflowCreated(payload)
	case domainworkflow.EventTypeWorkflowUpdated:
		return projectWorkflowUpdated(payload)
	case domainworkflow.EventTypeWorkflowDeleted:
		return projectWorkflowDeleted(payload)
	case domainstep.EventTypeStepCreated:
		return projectStepCreated(payload)
	case domainstep.EventTypeStepUpdated:
		return projectStepUpdated(payload)
	case domainstep.EventTypeStepDeleted:
		return projectStepDeleted(payload)
	case domainconnection.EventTypeConnectionCreated:
		return projectConnectionCreated(payload)
	case domainconnection.EventTypeConnectionDeleted:
		return projectConnectionDeleted(payload)
	case domainvariable.EventTypeVariableCreated:
		return projectVariableCreated(payload)
	case domainvariable.EventTypeVariableUpdated:
		return projectVariableUpdated(payload)
	case domainassertion.EventTypeAssertionCreated:
		return projectAssertionCreated(payload)
	case domainassertion.EventTypeAssertionUpdated:
		return projectAssertionUpdated(payload)
	case domainendpoint.EventTypeEndpointCreated:
		return projectEndpointCreated(payload)
	case domainendpoint.EventTypeEndpointUpdated:
		return projectEndpointUpdated(payload)
	case domainendpoint.EventTypeEndpointDeleted:
		return projectEndpointDeleted(payload)
	case domainendpoint.EventTypeEndpointImported:
		return projectEndpointImported(payload)
	case domainworkflowrun.EventTypeWorkflowRunStarted:
		return projectWorkflowRunStarted(payload)
	case domainworkflowrun.EventTypeWorkflowRunFinished:
		return projectWorkflowRunFinished(payload)
	case domainworkflowrun.EventTypeWorkflowRunCancelled:
		return projectWorkflowRunCancelled(payload)
	case domainworkflowrun.EventTypeWorkflowRunScheduledSkipped:
		return projectWorkflowRunScheduledSkipped(payload)
	case domainsteprun.EventTypeStepRunQueued:
		return projectStepRunQueued(payload)
	case domainsteprun.EventTypeStepRunStarted:
		return projectStepRunStarted(payload)
	case domainsteprun.EventTypeStepRunSucceeded:
		return projectStepRunSucceeded(payload)
	case domainsteprun.EventTypeStepRunFailed:
		return projectStepRunFailed(payload)
	default:
		return nil, nil
	}
}

func projectWorkflowCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflow.WorkflowCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionWorkflowCreated,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("Workflow %q created", evt.Name),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectWorkflowUpdated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflow.WorkflowUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionWorkflowUpdated,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("Workflow %q updated", evt.Name),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectWorkflowDeleted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflow.WorkflowDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionWorkflowDeleted,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		message:         "Workflow deleted",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectStepCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainstep.StepCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepCreated,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("Step %q created", evt.Name),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectStepUpdated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainstep.StepUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepUpdated,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("Step %q updated", evt.Name),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectStepDeleted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainstep.StepDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepDeleted,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelWarning,
		message:         "Step deleted",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectConnectionCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainconnection.ConnectionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	connectionID, err := uuid.Parse(evt.ConnectionID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionConnectionCreated,
		subjectType:     domainactivitylog.SubjectConnection,
		subjectID:       connectionID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		message:         "Connection created",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectConnectionDeleted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainconnection.ConnectionDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	connectionID, err := uuid.Parse(evt.ConnectionID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionConnectionDeleted,
		subjectType:     domainactivitylog.SubjectConnection,
		subjectID:       connectionID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		message:         "Connection deleted",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectVariableCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainvariable.VariableCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectVariable(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.VariableID, evt.ID, evt.EventType(),
		domainactivitylog.ActionVariableCreated, fmt.Sprintf("Variable %q created", evt.Key), evt.Timestamp, structPayload(evt))
}

func projectVariableUpdated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainvariable.VariableUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectVariable(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.VariableID, evt.ID, evt.EventType(),
		domainactivitylog.ActionVariableUpdated, fmt.Sprintf("Variable %q updated", evt.Key), evt.Timestamp, structPayload(evt))
}

func projectVariable(
	projectIDRaw, workflowIDRaw, stepIDRaw, variableIDRaw, eventIDRaw, eventType, action, message string,
	occurredAt time.Time, payload map[string]any,
) (*domainactivitylog.Entry, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, err
	}
	variableID, err := uuid.Parse(variableIDRaw)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return nil, err
	}
	var stepID *uuid.UUID
	if stepIDRaw != "" {
		parsed, err := uuid.Parse(stepIDRaw)
		if err != nil {
			return nil, err
		}
		stepID = &parsed
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectVariable,
		subjectID:       variableID,
		workflowID:      &workflowID,
		stepID:          stepID,
		level:           domainactivitylog.LevelInfo,
		message:         message,
		payload:         payload,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	}), nil
}

func projectAssertionCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainassertion.AssertionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectAssertionFromFields(
		evt.ProjectID, evt.WorkflowID, evt.StepID, evt.AssertionID, evt.ID, evt.EventType(),
		domainactivitylog.ActionAssertionCreated, "Assertion created", evt.Timestamp, structPayload(evt),
	)
}

func projectAssertionUpdated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainassertion.AssertionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectAssertionFromFields(
		evt.ProjectID, evt.WorkflowID, evt.StepID, evt.AssertionID, evt.ID, evt.EventType(),
		domainactivitylog.ActionAssertionUpdated, "Assertion updated", evt.Timestamp, structPayload(evt),
	)
}

func projectAssertionFromFields(
	projectIDRaw, workflowIDRaw, stepIDRaw, assertionIDRaw, eventIDRaw, eventType, action, message string,
	occurredAt time.Time, payload map[string]any,
) (*domainactivitylog.Entry, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, err
	}
	stepID, err := uuid.Parse(stepIDRaw)
	if err != nil {
		return nil, err
	}
	assertionID, err := uuid.Parse(assertionIDRaw)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectAssertion,
		subjectID:       assertionID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		message:         message,
		payload:         payload,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	}), nil
}

func projectEndpointCreated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainendpoint.EndpointCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointCreated, fmt.Sprintf("Endpoint %q created", evt.Name), evt.Timestamp, structPayload(evt))
}

func projectEndpointUpdated(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainendpoint.EndpointUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointUpdated, fmt.Sprintf("Endpoint %q updated", evt.Name), evt.Timestamp, structPayload(evt))
}

func projectEndpointDeleted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainendpoint.EndpointDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointDeleted, "Endpoint deleted", evt.Timestamp, structPayload(evt))
}

func projectEndpointImported(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainendpoint.EndpointImported
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, err := uuid.Parse(evt.ProjectID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionEndpointImported,
		subjectType:     domainactivitylog.SubjectEndpoint,
		subjectID:       projectID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("%d endpoints imported", evt.Count),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectEndpoint(
	projectIDRaw, endpointIDRaw, eventIDRaw, eventType, action, message string,
	occurredAt time.Time, payload map[string]any,
) (*domainactivitylog.Entry, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return nil, err
	}
	endpointID, err := uuid.Parse(endpointIDRaw)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectEndpoint,
		subjectID:       endpointID,
		level:           domainactivitylog.LevelInfo,
		message:         message,
		payload:         payload,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	}), nil
}

func projectWorkflowRunStarted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflowrun.WorkflowRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	actorType, actorUserID := actorFromTriggeredBy(evt.TriggeredBy, evt.TriggeredByUserID)
	return newEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunStarted,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		actorType:       actorType,
		actorUserID:     actorUserID,
		level:           domainactivitylog.LevelInfo,
		message:         fmt.Sprintf("Workflow run started (%s)", evt.TriggeredBy),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectWorkflowRunFinished(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflowrun.WorkflowRunFinished
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	level := domainactivitylog.LevelInfo
	message := fmt.Sprintf("Workflow run finished (%s)", evt.FinishType)
	if evt.FinishType == domainworkflowrun.FinishTypeFailed {
		level = domainactivitylog.LevelError
		message = "Workflow run failed"
	}
	if evt.FinishType == domainworkflowrun.FinishTypeCancelled {
		level = domainactivitylog.LevelWarning
		message = "Workflow run cancelled"
	}
	return newEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunFinished,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		level:           level,
		message:         message,
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectWorkflowRunCancelled(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflowrun.WorkflowRunCancelled
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunCancelled,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		level:           domainactivitylog.LevelWarning,
		message:         "Workflow run cancelled",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectWorkflowRunScheduledSkipped(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainworkflowrun.WorkflowRunScheduledSkipped
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunScheduledSkip,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		message:         fmt.Sprintf("Scheduled run skipped (%s)", evt.Reason),
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectStepRunQueued(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainsteprun.StepRunQueued
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	stepRunID, err := uuid.Parse(evt.StepRunID)
	if err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return nil, err
	}
	stepID, err := uuid.Parse(evt.StepID)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(evt.ID)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		action:          domainactivitylog.ActionStepRunQueued,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           domainactivitylog.LevelInfo,
		message:         "Step run queued",
		payload:         structPayload(evt),
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	}), nil
}

func projectStepRunStarted(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainsteprun.StepRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectStepRunFromFields(
		evt.ProjectID, evt.WorkflowID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunStarted,
		fmt.Sprintf("Step %q started (attempt %d)", evt.Name, evt.Attempt),
		domainactivitylog.LevelInfo, evt.Timestamp, structPayload(evt),
	)
}

func projectStepRunSucceeded(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainsteprun.StepRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	message := fmt.Sprintf("Step run succeeded (attempt %d)", evt.Attempt)
	if evt.ResponseSnapshot != nil {
		message = fmt.Sprintf("Step run succeeded — HTTP %d (attempt %d)", evt.ResponseSnapshot.Status, evt.Attempt)
	}
	return projectStepRunFromEvent(evt.ProjectID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunSucceeded, message, domainactivitylog.LevelInfo, evt.Timestamp, structPayload(evt))
}

func projectStepRunFailed(payload []byte) (*domainactivitylog.Entry, error) {
	var evt domainsteprun.StepRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	message := "Step run failed"
	if evt.Error != "" {
		message = fmt.Sprintf("Step run failed — %s", evt.Error)
	}
	return projectStepRunFromEvent(evt.ProjectID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunFailed, message, domainactivitylog.LevelError, evt.Timestamp, structPayload(evt))
}

func projectStepRunFromEvent(
	projectIDRaw, workflowRunIDRaw, stepIDRaw, stepRunIDRaw, eventIDRaw, eventType, action, message, level string,
	occurredAt time.Time, payload map[string]any,
) (*domainactivitylog.Entry, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(workflowRunIDRaw)
	if err != nil {
		return nil, err
	}
	stepID, err := uuid.Parse(stepIDRaw)
	if err != nil {
		return nil, err
	}
	stepRunID, err := uuid.Parse(stepRunIDRaw)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           level,
		message:         message,
		payload:         payload,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	}), nil
}

func projectStepRunFromFields(
	projectIDRaw, workflowIDRaw, workflowRunIDRaw, stepIDRaw, stepRunIDRaw, eventIDRaw, eventType, action, message, level string,
	occurredAt time.Time, payload map[string]any,
) (*domainactivitylog.Entry, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return nil, err
	}
	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return nil, err
	}
	workflowRunID, err := uuid.Parse(workflowRunIDRaw)
	if err != nil {
		return nil, err
	}
	stepID, err := uuid.Parse(stepIDRaw)
	if err != nil {
		return nil, err
	}
	stepRunID, err := uuid.Parse(stepRunIDRaw)
	if err != nil {
		return nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return nil, err
	}
	return newEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           level,
		message:         message,
		payload:         payload,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	}), nil
}

type newEntryParams struct {
	projectID       uuid.UUID
	action          string
	subjectType     string
	subjectID       uuid.UUID
	workflowID      *uuid.UUID
	workflowRunID   *uuid.UUID
	stepID          *uuid.UUID
	stepRunID       *uuid.UUID
	actorType       string
	actorUserID     *uuid.UUID
	level           string
	message         string
	payload         map[string]any
	sourceEventID   uuid.UUID
	sourceEventType string
	occurredAt      time.Time
}

func newEntry(p newEntryParams) *domainactivitylog.Entry {
	return &domainactivitylog.Entry{
		ID:              uuid.New(),
		ProjectID:       p.projectID,
		Action:          p.action,
		SubjectType:     p.subjectType,
		SubjectID:       p.subjectID,
		WorkflowID:      p.workflowID,
		WorkflowRunID:   p.workflowRunID,
		StepID:          p.stepID,
		StepRunID:       p.stepRunID,
		ActorType:       p.actorType,
		ActorUserID:     p.actorUserID,
		Level:           p.level,
		Message:         p.message,
		Payload:         p.payload,
		SourceEventID:   p.sourceEventID,
		SourceEventType: p.sourceEventType,
		OccurredAt:      p.occurredAt,
	}
}

func parseWorkflowScope(projectIDRaw, workflowIDRaw, eventIDRaw string) (uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, err := uuid.Parse(projectIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	workflowID, err := uuid.Parse(workflowIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	eventID, err := uuid.Parse(eventIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectID, workflowID, eventID, nil
}

func parseStepScope(projectIDRaw, workflowIDRaw, stepIDRaw, eventIDRaw string) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, error) {
	projectID, workflowID, eventID, err := parseWorkflowScope(projectIDRaw, workflowIDRaw, eventIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	stepID, err := uuid.Parse(stepIDRaw)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil, err
	}
	return projectID, workflowID, stepID, eventID, nil
}

func structPayload(v any) map[string]any {
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	return out
}

func actorFromTriggeredBy(triggeredBy string, userID *string) (string, *uuid.UUID) {
	switch domainworkflowrun.TriggeredBy(triggeredBy) {
	case domainworkflowrun.TriggeredByUser, domainworkflowrun.TriggeredByAPI:
		if userID != nil && *userID != "" {
			if parsed, err := uuid.Parse(*userID); err == nil {
				return domainactivitylog.ActorUser, &parsed
			}
		}
		return domainactivitylog.ActorAPI, nil
	case domainworkflowrun.TriggeredBySchedule:
		return domainactivitylog.ActorSchedule, nil
	case domainworkflowrun.TriggeredByCLI:
		return domainactivitylog.ActorCLI, nil
	case domainworkflowrun.TriggeredByWebhook:
		return domainactivitylog.ActorWebhook, nil
	default:
		return domainactivitylog.ActorSystem, nil
	}
}
