package workflow

import (
	"context"
	"errors"
	"time"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type CreateWorkflowCommand struct {
	UserID                uuid.UUID
	Name                  string
	Description           string
	ProjectID        uuid.UUID
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

type CreateWorkflowHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
	assert *cmdquota.AssertCreateAllowedHandler
}

func NewCreateWorkflowHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *CreateWorkflowHandler {
	return &CreateWorkflowHandler{repo: repo, outbox: outbox, assert: assert}
}

func (h *CreateWorkflowHandler) Handle(
	ctx context.Context,
	cmd CreateWorkflowCommand,
) (*domainworkflow.Workflow, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}

	if err := h.assert.AssertWorkflowCreate(ctx, cmd.UserID, cmd.ProjectID); err != nil {
		return nil, err
	}

	w, err := domainworkflow.NewWorkflow(domainworkflow.NewWorkflowParams{
		Name:                  cmd.Name,
		Description:           cmd.Description,
		ProjectID:        cmd.ProjectID,
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
	})
	if err != nil {
		return nil, err
	}

	err = h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, w); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(w.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, errors.New("failed to create workflow")
	}

	return w, nil
}
