package step

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type UpdateStepPositionCommand struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	WorkflowID     uuid.UUID
	Index          string
	Position       domainstep.Position
}

type UpdateStepPositionHandler struct {
	stepRepo domainstep.StepWriteRepository
	outbox   port.OutboxRepository
}

func NewUpdateStepPositionHandler(
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *UpdateStepPositionHandler {
	return &UpdateStepPositionHandler{
		stepRepo: stepRepo,
		outbox:   outbox,
	}
}

func (h *UpdateStepPositionHandler) Handle(
	ctx context.Context,
	cmd UpdateStepPositionCommand,
) (*domainstep.Step, error) {
	s, err := h.stepRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if s == nil || s.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if s.OrganizationID != cmd.OrganizationID || s.WorkflowID != cmd.WorkflowID {
		return nil, errors.New("step not found")
	}

	s.ApplyPositionUpdate(cmd.Index, cmd.Position)

	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, s.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	return s, nil
}
