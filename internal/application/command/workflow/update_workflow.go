package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type UpdateWorkflowCommand struct {
	ID                      uuid.UUID
	Name                    string
	Description             string
	Status                  domainworkflow.Status
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
}

type UpdateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewUpdateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *UpdateWorkflowHandler {
	return &UpdateWorkflowHandler{repo: repo, outbox: outbox}
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

		w.ApplyUpdate(domainworkflow.UpdateWorkflowParams{
			Name:                    cmd.Name,
			Description:             cmd.Description,
			Status:                  cmd.Status,
			ScheduleIntervalMinutes: cmd.ScheduleIntervalMinutes,
			Concurrency:             cmd.Concurrency,
			NotificationsEnabled:    cmd.NotificationsEnabled,
			NotifyOnSuccess:         cmd.NotifyOnSuccess,
			NotifyOnFailure:         cmd.NotifyOnFailure,
			NotifyOnCancel:          cmd.NotifyOnCancel,
		})

		if err := h.repo.Update(txCtx, w); err != nil {
			return errors.New("failed to update workflow")
		}
		return h.outbox.StoreEvents(txCtx, w.PullEvents())
	})
}
