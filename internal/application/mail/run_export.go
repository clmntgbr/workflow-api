package mail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"
)

const xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

type RunExportMailService struct {
	sender     port.MailSender
	appBaseURL string
}

func NewRunExportMailService(sender port.MailSender, appBaseURL string) *RunExportMailService {
	return &RunExportMailService{
		sender:     sender,
		appBaseURL: strings.TrimRight(appBaseURL, "/"),
	}
}

func (s *RunExportMailService) NotifyReady(
	ctx context.Context,
	workflow domainworkflow.WorkflowView,
	from, to time.Time,
	filename string,
	file []byte,
	recipients []string,
) error {
	if len(recipients) == 0 {
		return nil
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           recipients,
		Subject:      fmt.Sprintf(`Run history export for "%s" is ready`, workflow.Name),
		TemplateName: "run_export_ready",
		TemplateData: map[string]any{
			"Preheader":    fmt.Sprintf(`The run history export for "%s" is attached.`, workflow.Name),
			"WorkflowName": workflow.Name,
			"From":         formatDate(from),
			"To":           formatDate(to),
			"DashboardURL": s.dashboardURL(workflow),
		},
		Attachments: []port.MailAttachment{{
			Filename:    filename,
			ContentType: xlsxContentType,
			Bytes:       file,
		}},
	})
}

func (s *RunExportMailService) NotifyFailed(
	ctx context.Context,
	workflow domainworkflow.WorkflowView,
	reason string,
	recipients []string,
) error {
	if len(recipients) == 0 {
		return nil
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           recipients,
		Subject:      fmt.Sprintf(`Run history export for "%s" failed`, workflow.Name),
		TemplateName: "run_export_failed",
		TemplateData: map[string]any{
			"Preheader":    fmt.Sprintf(`We could not generate the run history export for "%s".`, workflow.Name),
			"WorkflowName": workflow.Name,
			"Error":        reason,
			"DashboardURL": s.dashboardURL(workflow),
		},
	})
}

func (s *RunExportMailService) dashboardURL(workflow domainworkflow.WorkflowView) string {
	if s.appBaseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/workflows/%s/runs", s.appBaseURL, workflow.ID)
}
