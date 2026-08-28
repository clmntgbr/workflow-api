package activitylog

import (
	"context"

	"go-api/internal/application/messaging"
	domainactivitylog "go-api/internal/domain/activitylog"
	domainstep "go-api/internal/domain/step"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type RecordHandler struct {
	repo            domainactivitylog.WriteRepository
	workflowRead    domainworkflow.WorkflowReadRepository
	workflowRunRead domainworkflowrun.WorkflowRunReadRepository
	stepRead        domainstep.StepReadRepository
	userRead        domainuser.UserReadRepository
}

func NewRecordHandler(
	repo domainactivitylog.WriteRepository,
	workflowRead domainworkflow.WorkflowReadRepository,
	workflowRunRead domainworkflowrun.WorkflowRunReadRepository,
	stepRead domainstep.StepReadRepository,
	userRead domainuser.UserReadRepository,
) *RecordHandler {
	return &RecordHandler{
		repo:            repo,
		workflowRead:    workflowRead,
		workflowRunRead: workflowRunRead,
		stepRead:        stepRead,
		userRead:        userRead,
	}
}

func (h *RecordHandler) ForEventType(eventType string) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		projected, err := Project(eventType, payload)
		if err != nil {
			return messaging.NonRetryable(err)
		}
		if projected == nil {
			return nil
		}
		if err := h.enrichProjectID(ctx, projected.entry); err != nil {
			return messaging.Retryable(err)
		}
		if projected.entry.ProjectID == uuid.Nil {
			return messaging.NonRetryable(errMissingProjectID)
		}
		applyPerformedByFromPayload(projected.entry, payload)
		if err := h.enrichHints(ctx, projected.entry, &projected.hints); err != nil {
			return messaging.Retryable(err)
		}
		projected.entry.Message = BuildMessage(projected.entry.Action, projected.hints)
		if err := h.repo.Save(ctx, projected.entry); err != nil {
			return messaging.Retryable(err)
		}
		return nil
	}
}

func (h *RecordHandler) enrichProjectID(ctx context.Context, entry *domainactivitylog.Entry) error {
	if entry.ProjectID != uuid.Nil {
		return nil
	}
	if entry.WorkflowID != nil {
		workflow, err := h.workflowRead.FindByID(ctx, *entry.WorkflowID)
		if err != nil {
			return err
		}
		if workflow != nil {
			entry.ProjectID = workflow.ProjectID
			return nil
		}
	}
	if entry.WorkflowRunID != nil {
		run, err := h.workflowRunRead.FindByID(ctx, *entry.WorkflowRunID)
		if err != nil {
			return err
		}
		if run != nil {
			entry.ProjectID = run.ProjectID
			if entry.WorkflowID == nil {
				entry.WorkflowID = &run.WorkflowID
			}
		}
	}
	return nil
}

type missingProjectIDError struct{}

func (missingProjectIDError) Error() string {
	return "unable to resolve projectId for activity log entry"
}

var errMissingProjectID error = missingProjectIDError{}
