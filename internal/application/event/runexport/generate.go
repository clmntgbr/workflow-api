package runexport

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"go-api/internal/application/mail"
	"go-api/internal/application/messaging"
	apprunexport "go-api/internal/application/runexport"
	domaininsight "go-api/internal/domain/insight"
	"go-api/internal/domain/port"
	domainrunexport "go-api/internal/domain/runexport"
	domainsteprun "go-api/internal/domain/steprun"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

const GenerateHandlerName = "RunExportGenerateHandler"

type insightsAllowedChecker interface {
	InsightsAllowedForProject(ctx context.Context, projectID uuid.UUID) (bool, error)
}

type GenerateHandler struct {
	jobs         domainrunexport.WriteRepository
	runs         domainworkflowrun.WorkflowRunReadRepository
	stepRuns     domainsteprun.StepRunReadRepository
	insights     domaininsight.InsightReadRepository
	workflows    domainworkflow.WorkflowReadRepository
	users        domainuser.UserReadRepository
	outbox       port.OutboxRepository
	insightsGate insightsAllowedChecker
	mail         *mail.RunExportMailService
}

func NewGenerateHandler(
	jobs domainrunexport.WriteRepository,
	runs domainworkflowrun.WorkflowRunReadRepository,
	stepRuns domainsteprun.StepRunReadRepository,
	insights domaininsight.InsightReadRepository,
	workflows domainworkflow.WorkflowReadRepository,
	users domainuser.UserReadRepository,
	outbox port.OutboxRepository,
	insightsGate insightsAllowedChecker,
	mailService *mail.RunExportMailService,
) *GenerateHandler {
	return &GenerateHandler{
		jobs:         jobs,
		runs:         runs,
		stepRuns:     stepRuns,
		insights:     insights,
		workflows:    workflows,
		users:        users,
		outbox:       outbox,
		insightsGate: insightsGate,
		mail:         mailService,
	}
}

func (h *GenerateHandler) Handle(ctx context.Context, payload []byte) error {
	var evt domainrunexport.RunExportRequested
	if err := json.Unmarshal(payload, &evt); err != nil {
		return messaging.NonRetryable(err)
	}

	jobID, err := uuid.Parse(evt.RunExportID)
	if err != nil {
		return messaging.NonRetryable(err)
	}

	job, err := h.jobs.GetByID(ctx, jobID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if job == nil {
		return messaging.Retryable(errors.New("run export not found"))
	}
	if job.Status == domainrunexport.StatusReady || job.Status == domainrunexport.StatusFailed {
		return nil
	}

	job.MarkProcessing()
	if err := h.jobs.Update(ctx, job); err != nil {
		return messaging.Retryable(err)
	}

	workflow, err := h.workflows.FindByID(ctx, job.WorkflowID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if workflow == nil {
		return h.fail(ctx, job, nil, "workflow not found")
	}

	recipients, err := h.recipientEmails(ctx, job.RequestedByUserID)
	if err != nil {
		return err
	}
	if len(recipients) == 0 {
		return h.fail(ctx, job, workflow, "requester has no deliverable email")
	}

	runs, err := h.runs.FindByWorkflowIDInRange(ctx, job.WorkflowID, domainworkflowrun.WorkflowRunRangeFilter{
		From: &job.DateRangeFrom,
		To:   &job.DateRangeTo,
	})
	if err != nil {
		return messaging.Retryable(err)
	}

	runIDs := make([]uuid.UUID, 0, len(runs))
	for _, run := range runs {
		runIDs = append(runIDs, run.ID)
	}

	stepRuns, err := h.stepRuns.FindByWorkflowRunIDs(ctx, runIDs)
	if err != nil {
		return messaging.Retryable(err)
	}

	withInsights, err := h.insightsGate.InsightsAllowedForProject(ctx, job.ProjectID)
	if err != nil {
		return messaging.Retryable(err)
	}

	var insights []domaininsight.InsightView
	if withInsights {
		stepRunIDs := make([]uuid.UUID, 0, len(stepRuns))
		for _, stepRun := range stepRuns {
			stepRunIDs = append(stepRunIDs, stepRun.ID)
		}
		insights, err = h.insights.FindByStepRunIDs(ctx, stepRunIDs)
		if err != nil {
			return messaging.Retryable(err)
		}
	}

	file, err := apprunexport.BuildXLSX(apprunexport.BuildInput{
		Runs:         runs,
		StepRuns:     stepRuns,
		Insights:     insights,
		WithInsights: withInsights,
	})
	if err != nil {
		if errors.Is(err, domainrunexport.ErrFileTooLarge) {
			return h.fail(ctx, job, workflow, domainrunexport.ErrFileTooLarge.Error())
		}
		return messaging.Retryable(err)
	}

	filename := apprunexport.ExportFilename(workflow.Name, job.DateRangeFrom, job.DateRangeTo)
	if err := h.mail.NotifyReady(ctx, *workflow, job.DateRangeFrom, job.DateRangeTo, filename, file, recipients); err != nil {
		return messaging.Retryable(err)
	}

	if err := h.finish(ctx, job, true, ""); err != nil {
		return messaging.Retryable(err)
	}

	log.Printf(
		"event handled %s eventId=%s runExportId=%s workflowId=%s",
		domainrunexport.EventTypeRunExportRequested,
		evt.ID,
		evt.RunExportID,
		evt.WorkflowID,
	)
	return nil
}

func (h *GenerateHandler) fail(
	ctx context.Context,
	job *domainrunexport.RunExport,
	workflow *domainworkflow.WorkflowView,
	reason string,
) error {
	if workflow != nil {
		recipients, err := h.recipientEmails(ctx, job.RequestedByUserID)
		if err != nil {
			return err
		}
		if err := h.mail.NotifyFailed(ctx, *workflow, reason, recipients); err != nil {
			return messaging.Retryable(err)
		}
	}
	if err := h.finish(ctx, job, false, reason); err != nil {
		return messaging.Retryable(err)
	}
	return nil
}

func (h *GenerateHandler) finish(
	ctx context.Context,
	job *domainrunexport.RunExport,
	ready bool,
	reason string,
) error {
	if ready {
		job.MarkReady()
	} else {
		job.MarkFailed(reason)
	}
	return h.jobs.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.jobs.Update(txCtx, job); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, job.PullEvents())
	})
}

func (h *GenerateHandler) recipientEmails(ctx context.Context, userID uuid.UUID) ([]string, error) {
	user, err := h.users.FindByID(ctx, userID)
	if err != nil {
		return nil, messaging.Retryable(err)
	}
	if user == nil {
		return nil, nil
	}
	return mail.RecipientsFromUser(user.Email, user.Banned), nil
}
