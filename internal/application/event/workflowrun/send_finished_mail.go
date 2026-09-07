package workflowrun

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"go-api/internal/application/mail"
	"go-api/internal/application/messaging"
	domainproject "go-api/internal/domain/project"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

const FinishedMailHandlerName = "WorkflowFinishedMailHandler"

type WorkflowFinishedMailHandler struct {
	workflows domainworkflow.WorkflowReadRepository
	runs      domainworkflowrun.WorkflowRunReadRepository
	projects  domainproject.ProjectReadRepository
	users     domainuser.UserReadRepository
	mail      *mail.WorkflowFinishedMailService
}

func NewWorkflowFinishedMailHandler(
	workflows domainworkflow.WorkflowReadRepository,
	runs domainworkflowrun.WorkflowRunReadRepository,
	projects domainproject.ProjectReadRepository,
	users domainuser.UserReadRepository,
	mailService *mail.WorkflowFinishedMailService,
) *WorkflowFinishedMailHandler {
	return &WorkflowFinishedMailHandler{
		workflows: workflows,
		runs:      runs,
		projects:  projects,
		users:     users,
		mail:      mailService,
	}
}

func (h *WorkflowFinishedMailHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainworkflowrun.WorkflowRunFinished
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	workflowID, err := uuid.Parse(evt.WorkflowID)
	if err != nil {
		return messaging.NonRetryable(err)
	}
	runID, err := uuid.Parse(evt.WorkflowRunID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	workflow, err := h.workflows.FindByID(ctx, workflowID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if workflow == nil {
		return messaging.NonRetryable(errWorkflowNotFoundForMail)
	}
	if !workflow.ShouldNotify(evt.FinishType) {
		return nil
	}

	run, err := h.runs.FindByID(ctx, runID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if run == nil {
		return messaging.NonRetryable(errWorkflowRunNotFoundForMail)
	}

	project, err := h.projects.FindByID(ctx, workflow.ProjectID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if project == nil {
		return messaging.NonRetryable(errProjectNotFoundForMail)
	}

	recipients, err := h.recipientEmails(ctx, project.MemberIDs)
	if err != nil {
		return messaging.Retryable(err)
	}

	if err := h.mail.Notify(ctx, *workflow, *run, evt.FinishType, recipients); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *WorkflowFinishedMailHandler) recipientEmails(ctx context.Context, memberIDs []uuid.UUID) ([]string, error) {
	emails := make([]string, 0, len(memberIDs))
	seen := make(map[string]struct{}, len(memberIDs))

	for _, memberID := range memberIDs {
		user, err := h.users.FindByID(ctx, memberID)
		if err != nil {
			return nil, err
		}
		if user == nil || user.Banned || user.Email == "" {
			continue
		}
		if _, exists := seen[user.Email]; exists {
			continue
		}
		seen[user.Email] = struct{}{}
		emails = append(emails, user.Email)
	}

	if len(emails) == 0 && len(memberIDs) > 0 {
		log.Printf("workflow finished mail: no deliverable member emails")
	}
	return emails, nil
}

var (
	errWorkflowNotFoundForMail    = errors.New("workflow not found for finished mail")
	errWorkflowRunNotFoundForMail = errors.New("workflow run not found for finished mail")
	errProjectNotFoundForMail     = errors.New("project not found for finished mail")
)
