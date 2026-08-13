package steprun

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type FailStepRunCommand struct {
	StepRunID uuid.UUID
	Error     string
	Response  *domainsteprun.ResponseSnapshot
}

type FailStepRunHandler struct {
	repo   domainsteprun.StepRunWriteRepository
	outbox port.OutboxRepository
}

func NewFailStepRunHandler(
	repo domainsteprun.StepRunWriteRepository,
	outbox port.OutboxRepository,
) *FailStepRunHandler {
	return &FailStepRunHandler{repo: repo, outbox: outbox}
}

func (h *FailStepRunHandler) Handle(ctx context.Context, cmd FailStepRunCommand) (*domainsteprun.StepRun, error) {
	if cmd.StepRunID == uuid.Nil {
		return nil, errors.New("stepRunId is required")
	}
	if cmd.Error == "" {
		return nil, errors.New("error is required")
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

	if err := run.MarkFailed(cmd.Error, cmd.Response); err != nil {
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
