package messaging

import (
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/event"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

func WithPerformedBy(events []event.DomainEvent, userID uuid.UUID) []event.DomainEvent {
	userRef := event.OptionalUserIDString(userID)
	if userRef == nil {
		return events
	}
	out := make([]event.DomainEvent, len(events))
	for i, evt := range events {
		out[i] = annotateEvent(evt, userRef)
	}
	return out
}

func annotateEvent(evt event.DomainEvent, userID *string) event.DomainEvent {
	switch e := evt.(type) {
	case domainworkflow.WorkflowCreated:
		e.PerformedByUserID = userID
		return e
	case domainworkflow.WorkflowUpdated:
		e.PerformedByUserID = userID
		return e
	case domainworkflow.WorkflowDeleted:
		e.PerformedByUserID = userID
		return e
	case domainstep.StepCreated:
		e.PerformedByUserID = userID
		return e
	case domainstep.StepUpdated:
		e.PerformedByUserID = userID
		return e
	case domainstep.StepPositionUpdated:
		e.PerformedByUserID = userID
		return e
	case domainstep.StepDeleted:
		e.PerformedByUserID = userID
		return e
	case domainconnection.ConnectionCreated:
		e.PerformedByUserID = userID
		return e
	case domainconnection.ConnectionDeleted:
		e.PerformedByUserID = userID
		return e
	case domainvariable.VariableCreated:
		e.PerformedByUserID = userID
		return e
	case domainvariable.VariableUpdated:
		e.PerformedByUserID = userID
		return e
	case domainvariable.VariableDeleted:
		e.PerformedByUserID = userID
		return e
	case domainassertion.AssertionCreated:
		e.PerformedByUserID = userID
		return e
	case domainassertion.AssertionUpdated:
		e.PerformedByUserID = userID
		return e
	case domainassertion.AssertionDeleted:
		e.PerformedByUserID = userID
		return e
	case domainendpoint.EndpointCreated:
		e.PerformedByUserID = userID
		return e
	case domainendpoint.EndpointUpdated:
		e.PerformedByUserID = userID
		return e
	case domainendpoint.EndpointDeleted:
		e.PerformedByUserID = userID
		return e
	case domainendpoint.EndpointImported:
		e.PerformedByUserID = userID
		return e
	case domainworkflowrun.WorkflowRunCancelled:
		e.PerformedByUserID = userID
		return e
	default:
		return evt
	}
}
