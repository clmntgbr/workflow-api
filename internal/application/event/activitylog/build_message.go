package activitylog

import (
	"fmt"
	"strings"

	domainactivitylog "go-api/internal/domain/activitylog"
	domainworkflowrun "go-api/internal/domain/workflowrun"
)

func BuildMessage(action string, h messageHints) string {
	return attachActor(action, buildMessageBody(action, h), h)
}

func buildMessageBody(action string, h messageHints) string {
	switch action {
	case domainactivitylog.ActionWorkflowCreated:
		return workflowLifecycleMessage("created", h)
	case domainactivitylog.ActionWorkflowUpdated:
		return workflowLifecycleMessage("updated", h)
	case domainactivitylog.ActionWorkflowActivated:
		return fmt.Sprintf("Workflow %s activated", quotedName(h.WorkflowName, "workflow"))
	case domainactivitylog.ActionWorkflowDeactivated:
		return fmt.Sprintf("Workflow %s deactivated", quotedName(h.WorkflowName, "workflow"))
	case domainactivitylog.ActionWorkflowDeleted:
		return fmt.Sprintf("Workflow %s deleted", quotedName(h.WorkflowName, "workflow"))

	case domainactivitylog.ActionStepCreated:
		return stepLifecycleMessage("added to", h)
	case domainactivitylog.ActionStepUpdated:
		return stepLifecycleMessage("updated in", h)
	case domainactivitylog.ActionStepPositionUpdated:
		return stepPositionMessage(h)
	case domainactivitylog.ActionStepDeleted:
		return fmt.Sprintf(
			"Step %s removed from workflow %s",
			quotedName(h.StepName, "step"),
			quotedName(h.WorkflowName, "workflow"),
		)

	case domainactivitylog.ActionConnectionCreated:
		return fmt.Sprintf(
			"Connected %s to %s in workflow %s",
			quotedName(h.SourceStepName, "source step"),
			quotedName(h.TargetStepName, "target step"),
			quotedName(h.WorkflowName, "workflow"),
		)
	case domainactivitylog.ActionConnectionDeleted:
		if h.SourceStepName != "" && h.TargetStepName != "" {
			return fmt.Sprintf(
				"Connection removed between %s and %s in workflow %s",
				quotedName(h.SourceStepName, "source step"),
				quotedName(h.TargetStepName, "target step"),
				quotedName(h.WorkflowName, "workflow"),
			)
		}
		return fmt.Sprintf(
			"Connection removed in workflow %s",
			quotedName(h.WorkflowName, "workflow"),
		)

	case domainactivitylog.ActionVariableCreated:
		return variableMessage("created", h)
	case domainactivitylog.ActionVariableUpdated:
		return variableMessage("updated", h)
	case domainactivitylog.ActionVariableDeleted:
		return variableMessage("deleted from", h)

	case domainactivitylog.ActionAssertionCreated:
		return assertionMessage("added", h)
	case domainactivitylog.ActionAssertionUpdated:
		return assertionMessage("updated", h)
	case domainactivitylog.ActionAssertionDeleted:
		return assertionMessage("deleted from", h)

	case domainactivitylog.ActionEndpointCreated:
		return endpointMessage("created", h)
	case domainactivitylog.ActionEndpointUpdated:
		return endpointMessage("updated", h)
	case domainactivitylog.ActionEndpointDeleted:
		return fmt.Sprintf("Endpoint %s deleted", quotedName(h.EndpointName, "endpoint"))
	case domainactivitylog.ActionEndpointImported:
		if h.ImportCount == 1 {
			return "1 endpoint imported"
		}
		return fmt.Sprintf("%d endpoints imported", h.ImportCount)

	case domainactivitylog.ActionWorkflowRunStarted:
		return fmt.Sprintf(
			"Workflow %s run started via %s",
			quotedName(h.WorkflowName, "workflow"),
			triggeredByLabel(h.TriggeredBy),
		)
	case domainactivitylog.ActionWorkflowRunFinished:
		return workflowRunFinishedMessage(h)
	case domainactivitylog.ActionWorkflowRunCancelled:
		return fmt.Sprintf("Workflow %s run cancelled", quotedName(h.WorkflowName, "workflow"))
	case domainactivitylog.ActionWorkflowRunScheduledSkip:
		return fmt.Sprintf(
			"Scheduled run of workflow %s skipped because %s",
			quotedName(h.WorkflowName, "workflow"),
			scheduledSkipReasonLabel(h.SkipReason),
		)

	case domainactivitylog.ActionStepRunQueued:
		return fmt.Sprintf(
			"Step %s queued in workflow %s run",
			quotedName(h.StepName, "step"),
			quotedName(h.WorkflowName, "workflow"),
		)
	case domainactivitylog.ActionStepRunStarted:
		return stepRunStartedMessage(h)
	case domainactivitylog.ActionStepRunSucceeded:
		return stepRunSucceededMessage(h)
	case domainactivitylog.ActionStepRunFailed:
		return stepRunFailedMessage(h)
	default:
		return action
	}
}

func attachActor(action, msg string, h messageHints) string {
	if h.ActorUserName == "" || !shouldAttachActor(action, h) {
		return msg
	}
	return msg + " by " + h.ActorUserName
}

func shouldAttachActor(action string, h messageHints) bool {
	switch action {
	case domainactivitylog.ActionStepRunQueued,
		domainactivitylog.ActionStepRunStarted,
		domainactivitylog.ActionStepRunSucceeded,
		domainactivitylog.ActionStepRunFailed,
		domainactivitylog.ActionWorkflowRunFinished,
		domainactivitylog.ActionWorkflowRunScheduledSkip:
		return false
	case domainactivitylog.ActionWorkflowRunStarted:
		switch h.TriggeredBy {
		case "user", "api":
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func workflowLifecycleMessage(verb string, h messageHints) string {
	msg := fmt.Sprintf("Workflow %s %s", quotedName(h.WorkflowName, "workflow"), verb)
	if h.WorkflowStatus != "" {
		msg += fmt.Sprintf(" with status %s", h.WorkflowStatus)
	}
	return msg
}

func stepLifecycleMessage(verb string, h messageHints) string {
	msg := fmt.Sprintf(
		"Step %s %s workflow %s",
		quotedName(h.StepName, "step"),
		verb,
		quotedName(h.WorkflowName, "workflow"),
	)
	if call := withHTTPCallPhrase(h.Method, h.URL); call != "" {
		msg += " " + call
	}
	return msg
}

func stepPositionMessage(h messageHints) string {
	msg := fmt.Sprintf(
		"Step %s moved on the canvas in workflow %s",
		quotedName(h.StepName, "step"),
		quotedName(h.WorkflowName, "workflow"),
	)
	if h.PositionX != 0 || h.PositionY != 0 {
		msg += fmt.Sprintf(" to position (%.0f, %.0f)", h.PositionX, h.PositionY)
	}
	return msg
}

func variableMessage(verb string, h messageHints) string {
	var msg string
	if verb == "deleted from" {
		msg = fmt.Sprintf(
			"Variable %q deleted from workflow %s",
			h.VariableKey,
			quotedName(h.WorkflowName, "workflow"),
		)
	} else {
		msg = fmt.Sprintf(
			"Variable %q %s on workflow %s",
			h.VariableKey,
			verb,
			quotedName(h.WorkflowName, "workflow"),
		)
	}
	if h.VariableKind != "" {
		msg += fmt.Sprintf(" as kind %s", h.VariableKind)
	}
	return msg
}

func assertionMessage(verb string, h messageHints) string {
	rule := strings.TrimSpace(fmt.Sprintf("%s %s", h.AssertionSource, h.AssertionOperator))
	if h.ExpectedValue != "" {
		rule = fmt.Sprintf("%s expected %q", rule, h.ExpectedValue)
	}
	if verb == "deleted from" {
		msg := fmt.Sprintf("Assertion deleted from step %s in workflow %s", quotedName(h.StepName, "step"), quotedName(h.WorkflowName, "workflow"))
		if rule != "" {
			msg += fmt.Sprintf(": %s", rule)
		}
		return msg
	}
	msg := fmt.Sprintf("Assertion %s on step %s", verb, quotedName(h.StepName, "step"))
	if rule != "" {
		msg += fmt.Sprintf(": %s", rule)
	}
	return msg
}

func endpointMessage(verb string, h messageHints) string {
	msg := fmt.Sprintf("Endpoint %s %s", quotedName(h.EndpointName, "endpoint"), verb)
	if call := withHTTPCallPhrase(h.Method, h.URL); call != "" {
		msg += " " + call
	}
	return msg
}

func workflowRunFinishedMessage(h messageHints) string {
	workflow := quotedName(h.WorkflowName, "workflow")
	switch domainworkflowrun.FinishType(h.FinishType) {
	case domainworkflowrun.FinishTypeFailed:
		msg := fmt.Sprintf("Workflow %s run failed", workflow)
		if h.Error != "" {
			msg += ": " + h.Error
		}
		return msg
	case domainworkflowrun.FinishTypeCancelled:
		return fmt.Sprintf("Workflow %s run cancelled", workflow)
	default:
		return fmt.Sprintf("Workflow %s run finished successfully", workflow)
	}
}

func stepRunStartedMessage(h messageHints) string {
	msg := fmt.Sprintf("Step %s started", quotedName(h.StepName, "step"))
	if call := withHTTPCallPhrase(h.Method, h.URL); call != "" {
		msg += " " + call
	}
	if attempt := onAttemptPhrase(h.Attempt, h.MaxAttempts); attempt != "" {
		msg += " " + attempt
	}
	return msg
}

func stepRunSucceededMessage(h messageHints) string {
	msg := fmt.Sprintf("Step %s succeeded", quotedName(h.StepName, "step"))
	if status := withHTTPStatusPhrase(h.StatusCode); status != "" {
		msg += " " + status
	}
	if attempt := onAttemptPhrase(h.Attempt, h.MaxAttempts); attempt != "" {
		msg += " " + attempt
	}
	return msg
}

func stepRunFailedMessage(h messageHints) string {
	msg := fmt.Sprintf("Step %s failed", quotedName(h.StepName, "step"))
	if status := withHTTPStatusPhrase(h.StatusCode); status != "" {
		msg += " " + status
	}
	if attempt := onAttemptPhrase(h.Attempt, h.MaxAttempts); attempt != "" {
		msg += " " + attempt
	}
	if h.Error != "" {
		msg += ": " + h.Error
	}
	return msg
}

func quotedName(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return fmt.Sprintf("%q", name)
	}
	return fallback
}
