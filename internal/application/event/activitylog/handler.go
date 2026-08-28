package activitylog

import (
	"context"

	"go-api/internal/application/messaging"
	domainactivitylog "go-api/internal/domain/activitylog"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type RecordHandler struct {
	repo            domainactivitylog.WriteRepository
	workflowRead    domainworkflow.WorkflowReadRepository
	workflowRunRead domainworkflowrun.WorkflowRunReadRepository
}

func NewRecordHandler(
	repo domainactivitylog.WriteRepository,
	workflowRead domainworkflow.WorkflowReadRepository,
	workflowRunRead domainworkflowrun.WorkflowRunReadRepository,
) *RecordHandler {
	return &RecordHandler{
		repo:            repo,
		workflowRead:    workflowRead,
		workflowRunRead: workflowRunRead,
	}
}

func (h *RecordHandler) ForEventType(eventType string) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		entry, err := Project(eventType, payload)
		if err != nil {
			return messaging.NonRetryable(err)
		}
		if entry == nil {
			return nil
		}
		if err := h.enrichProjectID(ctx, entry); err != nil {
			return messaging.Retryable(err)
		}
		if entry.ProjectID == uuid.Nil {
			return messaging.NonRetryable(errMissingProjectID)
		}
		if err := h.repo.Save(ctx, entry); err != nil {
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
