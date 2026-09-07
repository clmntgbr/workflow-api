package runexport

import (
	"context"
	"errors"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/messaging"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/port"
	domainrunexport "go-api/internal/domain/runexport"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type RequestRunExportCommand struct {
	UserID     uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	From       *time.Time
	To         *time.Time
}

type RequestRunExportHandler struct {
	jobs      domainrunexport.WriteRepository
	workflows domainworkflow.WorkflowWriteRepository
	outbox    port.OutboxRepository
	assert    *cmdquota.AssertCreateAllowedHandler
	retention querysubscription.RunHistoryCutoffResolver
}

func NewRequestRunExportHandler(
	jobs domainrunexport.WriteRepository,
	workflows domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
	retention querysubscription.RunHistoryCutoffResolver,
) *RequestRunExportHandler {
	return &RequestRunExportHandler{
		jobs:      jobs,
		workflows: workflows,
		outbox:    outbox,
		assert:    assert,
		retention: retention,
	}
}

func (h *RequestRunExportHandler) Handle(
	ctx context.Context,
	cmd RequestRunExportCommand,
) (*domainrunexport.RunExport, error) {
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if cmd.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}

	if err := h.assert.AssertDataExportAllowed(ctx, cmd.UserID, cmd.ProjectID); err != nil {
		return nil, err
	}

	workflow, err := h.workflows.GetByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to get workflow")
	}
	if workflow == nil || workflow.Status == domainworkflow.StatusDeleted {
		return nil, errors.New("workflow not found")
	}
	if workflow.ProjectID != cmd.ProjectID {
		return nil, errors.New("workflow not found")
	}

	from, to, err := h.resolveRange(ctx, cmd)
	if err != nil {
		return nil, err
	}

	job := domainrunexport.NewRunExport(domainrunexport.NewRunExportParams{
		WorkflowID:        cmd.WorkflowID,
		ProjectID:         cmd.ProjectID,
		RequestedByUserID: cmd.UserID,
		DateRangeFrom:     from,
		DateRangeTo:       to,
	})

	err = h.jobs.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.jobs.Save(txCtx, job); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(job.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, errors.New("failed to request run export")
	}
	return job, nil
}

func (h *RequestRunExportHandler) resolveRange(
	ctx context.Context,
	cmd RequestRunExportCommand,
) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	to := now
	if cmd.To != nil {
		to = cmd.To.UTC()
	}
	if to.After(now) {
		to = now
	}

	cutoff, err := h.retention.RunHistoryCutoff(ctx, cmd.UserID, cmd.ProjectID)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("failed to resolve run history retention")
	}

	from := querysubscription.ClampTimeFrom(cmd.From, cutoff)
	if from == nil {
		epoch := time.Unix(0, 0).UTC()
		from = &epoch
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, domainrunexport.ErrInvalidDateRange
	}
	return from.UTC(), to, nil
}
