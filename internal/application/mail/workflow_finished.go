package mail

import (
	"context"
	"fmt"
	"strings"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"
)

type WorkflowFinishedMailService struct {
	sender     port.MailSender
	appBaseURL string
}

func NewWorkflowFinishedMailService(sender port.MailSender, appBaseURL string) *WorkflowFinishedMailService {
	return &WorkflowFinishedMailService{
		sender:     sender,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

func (s *WorkflowFinishedMailService) Notify(
	ctx context.Context,
	workflow domainworkflow.WorkflowView,
	run domainworkflowrun.WorkflowRunView,
	finishType domainworkflowrun.FinishType,
	recipients []string,
) error {
	if !workflow.ShouldNotify(finishType) {
		return nil
	}
	if len(recipients) == 0 {
		return nil
	}

	subject, preheader := subjectFor(finishType, workflow.Name)
	return s.sender.Send(ctx, port.SendMailInput{
		To:           recipients,
		Subject:      subject,
		TemplateName: templateFor(finishType),
		TemplateData: map[string]any{
			"Preheader":    preheader,
			"WorkflowName": workflow.Name,
			"RunID":        run.ID.String(),
			"FinishType":   string(finishType),
			"StartedAt":    formatTime(run.StartedAt),
			"FinishedAt":   formatTime(run.FinishedAt),
			"DashboardURL": s.dashboardURL(workflow, run),
		},
	})
}

func templateFor(finishType domainworkflowrun.FinishType) string {
	switch finishType {
	case domainworkflowrun.FinishTypeSuccess:
		return "workflow_finished_success"
	case domainworkflowrun.FinishTypeFailed:
		return "workflow_finished_failed"
	case domainworkflowrun.FinishTypeCancelled:
		return "workflow_finished_cancelled"
	default:
		return ""
	}
}

func subjectFor(finishType domainworkflowrun.FinishType, workflowName string) (subject, preheader string) {
	switch finishType {
	case domainworkflowrun.FinishTypeSuccess:
		return fmt.Sprintf(`Workflow "%s" succeeded`, workflowName),
			fmt.Sprintf(`"%s" finished successfully.`, workflowName)
	case domainworkflowrun.FinishTypeFailed:
		return fmt.Sprintf(`Workflow "%s" failed`, workflowName),
			fmt.Sprintf(`"%s" did not complete successfully.`, workflowName)
	case domainworkflowrun.FinishTypeCancelled:
		return fmt.Sprintf(`Workflow "%s" was cancelled`, workflowName),
			fmt.Sprintf(`"%s" was cancelled before it finished.`, workflowName)
	default:
		return fmt.Sprintf(`Workflow "%s" finished`, workflowName),
			fmt.Sprintf(`"%s" has finished.`, workflowName)
	}
}

func (s *WorkflowFinishedMailService) dashboardURL(
	workflow domainworkflow.WorkflowView,
	run domainworkflowrun.WorkflowRunView,
) string {
	if s.appBaseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/workflows/%s/runs/%s", s.appBaseURL, workflow.ID, run.ID)
}
