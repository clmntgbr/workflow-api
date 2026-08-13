package steprun

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type StartStepRunCommand struct {
	StepRunID uuid.UUID
}

type StartStepRunHandler struct {
	repo   domainsteprun.StepRunWriteRepository
	outbox port.OutboxRepository
}

func NewStartStepRunHandler(
	repo domainsteprun.StepRunWriteRepository,
	outbox port.OutboxRepository,
) *StartStepRunHandler {
	return &StartStepRunHandler{repo: repo, outbox: outbox}
}

func (h *StartStepRunHandler) Handle(ctx context.Context, cmd StartStepRunCommand) (*domainsteprun.StepRun, error) {
	if cmd.StepRunID == uuid.Nil {
		return nil, errors.New("stepRunId is required")
	}

	run, err := h.repo.GetByID(ctx, cmd.StepRunID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, domainsteprun.ErrNotFound
	}
	if run.Status.IsTerminal() {
		return run, nil
	}
	if run.Status == domainsteprun.StatusRunning {
		return run, nil
	}

	if err := run.MarkStarted(); err != nil {
		return nil, err
	}

	err = h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Update(txCtx, run); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, run.PullEvents())
	})
	if err != nil {
		return nil, err
	}
	return run, nil
}
