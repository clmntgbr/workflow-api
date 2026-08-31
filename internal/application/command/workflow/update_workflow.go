package workflow

import (
	"context"
	"errors"
	"time"

	"go-api/internal/application/messaging"
	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type UpdateWorkflowCommand struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	Name                  string
	Description           string
	Status                domainworkflow.Status
	ScheduleType          domainworkflow.ScheduleType
	ScheduleIntervalValue int
	ScheduleIntervalUnit  domainworkflow.ScheduleUnit
	ScheduleAt            *time.Time
	ScheduleTimezone      string
	Concurrency           int
	NotificationsEnabled  bool
	NotifyOnSuccess       bool
	NotifyOnFailure       bool
	NotifyOnCancel        bool
}

type UpdateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
	assert *cmdquota.AssertCreateAllowedHandler
}

func NewUpdateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *UpdateWorkflowHandler {
	return &UpdateWorkflowHandler{repo: repo, outbox: outbox, assert: assert}
}

func (h *UpdateWorkflowHandler) Handle(ctx context.Context, cmd UpdateWorkflowCommand) error {
	if cmd.Name == "" {
		return errors.New("name is required")
	}
	if !cmd.Status.Valid() {
		return errors.New("invalid status")
	}
	if cmd.Status == domainworkflow.StatusDeleted {
		return errors.New("use delete to mark a workflow as deleted")
	}

	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		w, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get workflow")
		}
		if w == nil || w.Status == domainworkflow.StatusDeleted {
			return errors.New("workflow not found")
		}

		if err := h.assert.AssertScheduleInterval(
			txCtx,
			cmd.UserID,
			w.ProjectID,
			cmd.ScheduleType,
			cmd.ScheduleIntervalValue,
			cmd.ScheduleIntervalUnit,
		); err != nil {
			return err
		}

		if err := w.ApplyUpdate(domainworkflow.UpdateWorkflowParams{
			Name:                  cmd.Name,
			Description:           cmd.Description,
			Status:                cmd.Status,
			ScheduleType:          cmd.ScheduleType,
			ScheduleIntervalValue: cmd.ScheduleIntervalValue,
			ScheduleIntervalUnit:  cmd.ScheduleIntervalUnit,
			ScheduleAt:            cmd.ScheduleAt,
			ScheduleTimezone:      cmd.ScheduleTimezone,
			Concurrency:           cmd.Concurrency,
			NotificationsEnabled:  cmd.NotificationsEnabled,
			NotifyOnSuccess:       cmd.NotifyOnSuccess,
			NotifyOnFailure:       cmd.NotifyOnFailure,
			NotifyOnCancel:        cmd.NotifyOnCancel,
		}); err != nil {
			return err
		}

		if err := h.repo.Update(txCtx, w); err != nil {
			return errors.New("failed to update workflow")
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(w.PullEvents(), cmd.UserID))
	})
}
