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

type UpdateConditionStepCommand struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	WorkflowID  uuid.UUID
	ProjectID   uuid.UUID
	Name        string
	Description string
	Expression  string
}

type UpdateConditionStepHandler struct {
	stepRepo domainstep.StepWriteRepository
	outbox   port.OutboxRepository
}

func NewUpdateConditionStepHandler(
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *UpdateConditionStepHandler {
	return &UpdateConditionStepHandler{
		stepRepo: stepRepo,
		outbox:   outbox,
	}
}

func (h *UpdateConditionStepHandler) Handle(
	ctx context.Context,
	cmd UpdateConditionStepCommand,
) (*domainstep.Step, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, errors.New("name is required")
	}
	if strings.TrimSpace(cmd.Expression) == "" {
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
	if s.Type != domainstep.TypeCondition {
		return nil, errors.New("step is not a condition step")
	}

	if err := s.ApplyConditionConfigUpdate(domainstep.UpdateConditionStepConfigParams{
		Name:        strings.TrimSpace(cmd.Name),
		Description: strings.TrimSpace(cmd.Description),
		Expression:  cmd.Expression,
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
