package steprun

import (
	"context"
	"errors"

	domainassertion "go-api/internal/domain/assertion"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type IncrementStepRunAttemptCommand struct {
	StepRunID        uuid.UUID
	Response         *domainsteprun.ResponseSnapshot
	Error            string
	AssertionsResult []domainassertion.Result
}

type IncrementStepRunAttemptHandler struct {
	repo domainsteprun.StepRunWriteRepository
}

func NewIncrementStepRunAttemptHandler(
	repo domainsteprun.StepRunWriteRepository,
) *IncrementStepRunAttemptHandler {
	return &IncrementStepRunAttemptHandler{repo: repo}
}

func (h *IncrementStepRunAttemptHandler) Handle(
	ctx context.Context,
	cmd IncrementStepRunAttemptCommand,
) (*domainsteprun.StepRun, error) {
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

	if err := run.IncrementAttempt(); err != nil {
		return nil, err
	}
	if cmd.Response != nil {
		normalized := cmd.Response.Normalized()
		run.ResponseSnapshot = &normalized
	}
	if cmd.AssertionsResult == nil {
		run.AssertionsResult = []domainassertion.Result{}
	} else {
		run.AssertionsResult = cmd.AssertionsResult
	}
	run.Error = cmd.Error

	if err := h.repo.Update(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}
