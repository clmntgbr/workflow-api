package activitylog

import (
	"encoding/json"
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

func Project(eventType string, payload []byte) (*projectedLog, error) {
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
	case domainstep.EventTypeStepPositionUpdated:
		return projectStepPositionUpdated(payload)
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
	case domainvariable.EventTypeVariableDeleted:
		return projectVariableDeleted(payload)
	case domainassertion.EventTypeAssertionCreated:
		return projectAssertionCreated(payload)
	case domainassertion.EventTypeAssertionUpdated:
		return projectAssertionUpdated(payload)
	case domainassertion.EventTypeAssertionDeleted:
		return projectAssertionDeleted(payload)
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

func projectWorkflowCreated(payload []byte) (*projectedLog, error) {
	var evt domainworkflow.WorkflowCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionWorkflowCreated,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{WorkflowName: evt.Name, WorkflowStatus: evt.Status},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectWorkflowUpdated(payload []byte) (*projectedLog, error) {
	var evt domainworkflow.WorkflowUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	action := domainactivitylog.ActionWorkflowUpdated
	switch evt.UpdateReason {
	case domainworkflow.WorkflowUpdateReasonActivated:
		action = domainactivitylog.ActionWorkflowActivated
	case domainworkflow.WorkflowUpdateReasonDeactivated:
		action = domainactivitylog.ActionWorkflowDeactivated
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{WorkflowName: evt.Name, WorkflowStatus: evt.Status},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectWorkflowDeleted(payload []byte) (*projectedLog, error) {
	var evt domainworkflow.WorkflowDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, eventID, err := parseWorkflowScope(evt.ProjectID, evt.WorkflowID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionWorkflowDeleted,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepCreated(payload []byte) (*projectedLog, error) {
	var evt domainstep.StepCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepCreated,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{StepName: evt.Name, Method: evt.Method, URL: evt.URL},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepUpdated(payload []byte) (*projectedLog, error) {
	var evt domainstep.StepUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepUpdated,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{StepName: evt.Name, Method: evt.Method, URL: evt.URL},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepPositionUpdated(payload []byte) (*projectedLog, error) {
	var evt domainstep.StepPositionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepPositionUpdated,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{
			StepName: evt.Name,
			PositionX: evt.Position.X,
			PositionY: evt.Position.Y,
		},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepDeleted(payload []byte) (*projectedLog, error) {
	var evt domainstep.StepDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	projectID, workflowID, stepID, eventID, err := parseStepScope(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.ID)
	if err != nil {
		return nil, err
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionStepDeleted,
		subjectType:     domainactivitylog.SubjectStep,
		subjectID:       stepID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           domainactivitylog.LevelWarning,
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectConnectionCreated(payload []byte) (*projectedLog, error) {
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
	sourceStepID, _ := uuid.Parse(evt.SourceStepID)
	targetStepID, _ := uuid.Parse(evt.TargetStepID)
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionConnectionCreated,
		subjectType:     domainactivitylog.SubjectConnection,
		subjectID:       connectionID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelInfo,
		hints: messageHints{
			SourceStepID: sourceStepID,
			TargetStepID: targetStepID,
		},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectConnectionDeleted(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionConnectionDeleted,
		subjectType:     domainactivitylog.SubjectConnection,
		subjectID:       connectionID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectVariableCreated(payload []byte) (*projectedLog, error) {
	var evt domainvariable.VariableCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectVariable(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.VariableID, evt.ID, evt.EventType(),
		domainactivitylog.ActionVariableCreated,
		evt.Timestamp,
		messageHints{VariableKey: evt.Key, VariableKind: evt.Kind})
}

func projectVariableUpdated(payload []byte) (*projectedLog, error) {
	var evt domainvariable.VariableUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectVariable(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.VariableID, evt.ID, evt.EventType(),
		domainactivitylog.ActionVariableUpdated,
		evt.Timestamp,
		messageHints{VariableKey: evt.Key, VariableKind: evt.Kind})
}

func projectVariableDeleted(payload []byte) (*projectedLog, error) {
	var evt domainvariable.VariableDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectVariable(evt.ProjectID, evt.WorkflowID, evt.StepID, evt.VariableID, evt.ID, evt.EventType(),
		domainactivitylog.ActionVariableDeleted,
		evt.Timestamp,
		messageHints{VariableKey: evt.Key, VariableKind: evt.Kind})
}

func projectVariable(
	projectIDRaw, workflowIDRaw, stepIDRaw, variableIDRaw, eventIDRaw, eventType, action string,
	occurredAt time.Time, hints messageHints,
) (*projectedLog, error) {
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
	level := domainactivitylog.LevelInfo
	if action == domainactivitylog.ActionVariableDeleted {
		level = domainactivitylog.LevelWarning
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectVariable,
		subjectID:       variableID,
		workflowID:      &workflowID,
		stepID:          stepID,
		level:           level,
		hints:           hints,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	})
}

func projectAssertionCreated(payload []byte) (*projectedLog, error) {
	var evt domainassertion.AssertionCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectAssertionFromFields(
		evt.ProjectID, evt.WorkflowID, evt.StepID, evt.AssertionID, evt.ID, evt.EventType(),
		domainactivitylog.ActionAssertionCreated,
		evt.Timestamp,
		messageHints{
			AssertionSource:   evt.Source,
			AssertionOperator: evt.Operator,
			ExpectedValue:     evt.ExpectedValue,
		},
	)
}

func projectAssertionUpdated(payload []byte) (*projectedLog, error) {
	var evt domainassertion.AssertionUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectAssertionFromFields(
		evt.ProjectID, evt.WorkflowID, evt.StepID, evt.AssertionID, evt.ID, evt.EventType(),
		domainactivitylog.ActionAssertionUpdated,
		evt.Timestamp,
		messageHints{
			AssertionSource:   evt.Source,
			AssertionOperator: evt.Operator,
			ExpectedValue:     evt.ExpectedValue,
		},
	)
}

func projectAssertionDeleted(payload []byte) (*projectedLog, error) {
	var evt domainassertion.AssertionDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectAssertionFromFields(
		evt.ProjectID, evt.WorkflowID, evt.StepID, evt.AssertionID, evt.ID, evt.EventType(),
		domainactivitylog.ActionAssertionDeleted,
		evt.Timestamp,
		messageHints{
			AssertionSource:   evt.Source,
			AssertionOperator: evt.Operator,
			ExpectedValue:     evt.ExpectedValue,
		},
	)
}

func projectAssertionFromFields(
	projectIDRaw, workflowIDRaw, stepIDRaw, assertionIDRaw, eventIDRaw, eventType, action string,
	occurredAt time.Time, hints messageHints,
) (*projectedLog, error) {
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
	level := domainactivitylog.LevelInfo
	if action == domainactivitylog.ActionAssertionDeleted {
		level = domainactivitylog.LevelWarning
	}
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectAssertion,
		subjectID:       assertionID,
		workflowID:      &workflowID,
		stepID:          &stepID,
		level:           level,
		hints:           hints,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	})
}

func projectEndpointCreated(payload []byte) (*projectedLog, error) {
	var evt domainendpoint.EndpointCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointCreated,
		evt.Timestamp,
		messageHints{EndpointName: evt.Name, Method: evt.Method, URL: evt.URL})
}

func projectEndpointUpdated(payload []byte) (*projectedLog, error) {
	var evt domainendpoint.EndpointUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointUpdated,
		evt.Timestamp,
		messageHints{EndpointName: evt.Name, Method: evt.Method, URL: evt.URL})
}

func projectEndpointDeleted(payload []byte) (*projectedLog, error) {
	var evt domainendpoint.EndpointDeleted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	return projectEndpoint(evt.ProjectID, evt.EndpointID, evt.ID, evt.EventType(),
		domainactivitylog.ActionEndpointDeleted,
		evt.Timestamp,
		messageHints{})
}

func projectEndpointImported(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          domainactivitylog.ActionEndpointImported,
		subjectType:     domainactivitylog.SubjectEndpoint,
		subjectID:       projectID,
		level:           domainactivitylog.LevelInfo,
		hints:           messageHints{ImportCount: evt.Count},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectEndpoint(
	projectIDRaw, endpointIDRaw, eventIDRaw, eventType, action string,
	occurredAt time.Time, hints messageHints,
) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectEndpoint,
		subjectID:       endpointID,
		level:           domainactivitylog.LevelInfo,
		hints:           hints,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	})
}

func projectWorkflowRunStarted(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunStarted,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		actorType:       actorType,
		actorUserID:     actorUserID,
		level:           domainactivitylog.LevelInfo,
		hints:           messageHints{TriggeredBy: evt.TriggeredBy},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectWorkflowRunFinished(payload []byte) (*projectedLog, error) {
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
	if evt.FinishType == domainworkflowrun.FinishTypeFailed {
		level = domainactivitylog.LevelError
	}
	if evt.FinishType == domainworkflowrun.FinishTypeCancelled {
		level = domainactivitylog.LevelWarning
	}
	return finalizeEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunFinished,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		level:           level,
		hints: messageHints{
			FinishType: string(evt.FinishType),
			Error:      evt.Error,
		},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectWorkflowRunCancelled(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunCancelled,
		subjectType:     domainactivitylog.SubjectWorkflowRun,
		subjectID:       workflowRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		level:           domainactivitylog.LevelWarning,
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectWorkflowRunScheduledSkipped(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		action:          domainactivitylog.ActionWorkflowRunScheduledSkip,
		subjectType:     domainactivitylog.SubjectWorkflow,
		subjectID:       workflowID,
		workflowID:      &workflowID,
		level:           domainactivitylog.LevelWarning,
		hints:           messageHints{SkipReason: evt.Reason},
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepRunQueued(payload []byte) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		action:          domainactivitylog.ActionStepRunQueued,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           domainactivitylog.LevelInfo,
		sourceEventID:   eventID,
		sourceEventType: evt.EventType(),
		occurredAt:      evt.Timestamp,
	})
}

func projectStepRunStarted(payload []byte) (*projectedLog, error) {
	var evt domainsteprun.StepRunStarted
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	maxAttempts := evt.RetryCount
	if !evt.RetryOnFailure || maxAttempts <= 0 {
		maxAttempts = 0
	}
	return projectStepRunFromFields(
		evt.ProjectID, evt.WorkflowID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunStarted,
		domainactivitylog.LevelInfo,
		evt.Timestamp,
		messageHints{
			StepName:    evt.Name,
			Method:      evt.Method,
			URL:         evt.URL,
			Attempt:     evt.Attempt,
			MaxAttempts: maxAttempts,
		},
	)
}

func projectStepRunSucceeded(payload []byte) (*projectedLog, error) {
	var evt domainsteprun.StepRunSucceeded
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	hints := messageHints{Attempt: evt.Attempt}
	if evt.ResponseSnapshot != nil {
		hints.StatusCode = evt.ResponseSnapshot.Status
	}
	return projectStepRunFromEvent(evt.ProjectID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunSucceeded, domainactivitylog.LevelInfo, evt.Timestamp, hints)
}

func projectStepRunFailed(payload []byte) (*projectedLog, error) {
	var evt domainsteprun.StepRunFailed
	if err := json.Unmarshal(payload, &evt); err != nil {
		return nil, err
	}
	hints := messageHints{Attempt: evt.Attempt, Error: evt.Error}
	if evt.ResponseSnapshot != nil {
		hints.StatusCode = evt.ResponseSnapshot.Status
	}
	return projectStepRunFromEvent(evt.ProjectID, evt.WorkflowRunID, evt.StepID, evt.StepRunID, evt.ID, evt.EventType(),
		domainactivitylog.ActionStepRunFailed, domainactivitylog.LevelError, evt.Timestamp, hints)
}

func projectStepRunFromEvent(
	projectIDRaw, workflowRunIDRaw, stepIDRaw, stepRunIDRaw, eventIDRaw, eventType, action, level string,
	occurredAt time.Time, hints messageHints,
) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           level,
		hints:           hints,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	})
}

func projectStepRunFromFields(
	projectIDRaw, workflowIDRaw, workflowRunIDRaw, stepIDRaw, stepRunIDRaw, eventIDRaw, eventType, action, level string,
	occurredAt time.Time, hints messageHints,
) (*projectedLog, error) {
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
	return finalizeEntry(newEntryParams{
		projectID:       projectID,
		action:          action,
		subjectType:     domainactivitylog.SubjectStepRun,
		subjectID:       stepRunID,
		workflowID:      &workflowID,
		workflowRunID:   &workflowRunID,
		stepID:          &stepID,
		stepRunID:       &stepRunID,
		level:           level,
		hints:           hints,
		sourceEventID:   eventID,
		sourceEventType: eventType,
		occurredAt:      occurredAt,
	})
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
	hints           messageHints
	sourceEventID   uuid.UUID
	sourceEventType string
	occurredAt      time.Time
}

func finalizeEntry(p newEntryParams) (*projectedLog, error) {
	return &projectedLog{
		entry: &domainactivitylog.Entry{
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
			SourceEventID:   p.sourceEventID,
			SourceEventType: p.sourceEventType,
			OccurredAt:      p.occurredAt,
		},
		hints: p.hints,
	}, nil
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
