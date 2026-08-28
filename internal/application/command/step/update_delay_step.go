package step

import (
	"context"
	"errors"
	"strings"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type UpdateDelayStepCommand struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	WorkflowID           uuid.UUID
	ProjectID            uuid.UUID
	Name                 string
	Description          string
	DelayDurationSeconds int
}

type UpdateDelayStepHandler struct {
	stepRepo domainstep.StepWriteRepository
	outbox   port.OutboxRepository
}

func NewUpdateDelayStepHandler(
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *UpdateDelayStepHandler {
	return &UpdateDelayStepHandler{
		stepRepo: stepRepo,
		outbox:   outbox,
	}
}

func (h *UpdateDelayStepHandler) Handle(
	ctx context.Context,
	cmd UpdateDelayStepCommand,
) (*domainstep.Step, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, errors.New("name is required")
	}
	if cmd.DelayDurationSeconds <= 0 {
		return nil, domainstep.ErrInvalidStepTypeConfig
	}

	s, err := h.stepRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if s == nil || s.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if s.ProjectID != cmd.ProjectID || s.WorkflowID != cmd.WorkflowID {
		return nil, errors.New("step not found")
	}
	if s.Type != domainstep.TypeDelay {
		return nil, errors.New("step is not a delay step")
	}

	if err := s.ApplyDelayConfigUpdate(domainstep.UpdateDelayStepConfigParams{
		Name:                 strings.TrimSpace(cmd.Name),
		Description:          strings.TrimSpace(cmd.Description),
		DelayDurationSeconds: cmd.DelayDurationSeconds,
	}); err != nil {
		return nil, err
	}

	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(s.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	return s, nil
}
