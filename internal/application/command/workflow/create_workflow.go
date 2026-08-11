package workflow

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type CreateWorkflowCommand struct {
	Name                    string
	Description             string
	OrganizationID          uuid.UUID
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
}

type CreateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewCreateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *CreateWorkflowHandler {
	return &CreateWorkflowHandler{repo: repo, outbox: outbox}
}

func (h *CreateWorkflowHandler) Handle(
	ctx context.Context,
	cmd CreateWorkflowCommand,
) (*domainworkflow.Workflow, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}

	w := domainworkflow.NewWorkflow(domainworkflow.NewWorkflowParams{
		Name:                    cmd.Name,
		Description:             cmd.Description,
		OrganizationID:          cmd.OrganizationID,
		ScheduleIntervalMinutes: cmd.ScheduleIntervalMinutes,
		Concurrency:             cmd.Concurrency,
		NotificationsEnabled:    cmd.NotificationsEnabled,
		NotifyOnSuccess:         cmd.NotifyOnSuccess,
		NotifyOnFailure:         cmd.NotifyOnFailure,
		NotifyOnCancel:          cmd.NotifyOnCancel,
	})

	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, w); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, w.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create workflow")
	}

	return w, nil
}
