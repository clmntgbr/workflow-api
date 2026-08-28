package activitylog

import (
	"go-api/internal/application/event/dedup"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"
)

const handlerName = "record_activity_log"

func Register(reg *registry.HandlerRegistry, dedupRepo port.ProcessedEventRepository, handler *RecordHandler) {
	eventTypes := []string{
		domainworkflow.EventTypeWorkflowCreated,
		domainworkflow.EventTypeWorkflowUpdated,
		domainworkflow.EventTypeWorkflowDeleted,
		domainstep.EventTypeStepCreated,
		domainstep.EventTypeStepUpdated,
		domainstep.EventTypeStepDeleted,
		domainconnection.EventTypeConnectionCreated,
		domainconnection.EventTypeConnectionDeleted,
		domainvariable.EventTypeVariableCreated,
		domainvariable.EventTypeVariableUpdated,
		domainassertion.EventTypeAssertionCreated,
		domainassertion.EventTypeAssertionUpdated,
		domainendpoint.EventTypeEndpointCreated,
		domainendpoint.EventTypeEndpointUpdated,
		domainendpoint.EventTypeEndpointDeleted,
		domainendpoint.EventTypeEndpointImported,
		domainworkflowrun.EventTypeWorkflowRunStarted,
		domainworkflowrun.EventTypeWorkflowRunFinished,
		domainworkflowrun.EventTypeWorkflowRunCancelled,
		domainworkflowrun.EventTypeWorkflowRunScheduledSkipped,
		domainsteprun.EventTypeStepRunQueued,
		domainsteprun.EventTypeStepRunStarted,
		domainsteprun.EventTypeStepRunSucceeded,
		domainsteprun.EventTypeStepRunFailed,
	}
	for _, eventType := range eventTypes {
		reg.Register(eventType, dedup.With(dedupRepo, handlerName, handler.ForEventType(eventType)))
	}
}
