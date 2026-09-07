package steprun

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

const (
	staleStepRunError     = "step run exceeded its execution budget"
	staleWorkflowRunError = "a step run stalled and was marked failed"
)

type FailStaleStepRunsHandler struct {
	stepRunRepo   domainsteprun.StepRunWriteRepository
	runRepo       domainworkflowrun.WorkflowRunWriteRepository
	outbox        port.OutboxRepository
	pendingMaxAge time.Duration
	grace         time.Duration
	maxBatches    int
}

func NewFailStaleStepRunsHandler(
	stepRunRepo domainsteprun.StepRunWriteRepository,
	runRepo domainworkflowrun.WorkflowRunWriteRepository,
	outbox port.OutboxRepository,
	pendingMaxAge time.Duration,
	grace time.Duration,
	maxBatches int,
) *FailStaleStepRunsHandler {
	if maxBatches <= 0 {
		maxBatches = 50
	}
	return &FailStaleStepRunsHandler{
		stepRunRepo:   stepRunRepo,
		runRepo:       runRepo,
		outbox:        outbox,
		pendingMaxAge: pendingMaxAge,
		grace:         grace,
		maxBatches:    maxBatches,
	}
}

func (h *FailStaleStepRunsHandler) Handle(ctx context.Context, now time.Time, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}

	failed := 0
	for batch := 0; batch < h.maxBatches; batch++ {
		n, err := h.handleBatch(ctx, now, limit)
		if err != nil {
			return failed, err
		}
		if n == 0 {
			break
		}
		failed += n
	}
	return failed, nil
}

func (h *FailStaleStepRunsHandler) handleBatch(ctx context.Context, now time.Time, limit int) (int, error) {
	candidates, err := h.stepRunRepo.FindActiveNonDelay(ctx, now, h.pendingMaxAge, h.grace, limit)
	if err != nil {
		return 0, err
	}

	failed := 0
	seen := make(map[uuid.UUID]struct{})
	for _, candidate := range candidates {
		if candidate == nil || !candidate.IsStale(now, h.pendingMaxAge, h.grace) {
			continue
		}
		if _, ok := seen[candidate.WorkflowRunID]; ok {
			continue
		}
		seen[candidate.WorkflowRunID] = struct{}{}

		marked, err := h.failWorkflowRun(ctx, candidate.WorkflowRunID, now)
		if err != nil {
			return failed, err
		}
		if marked {
			failed++
		}
	}
	return failed, nil
}

func (h *FailStaleStepRunsHandler) failWorkflowRun(
	ctx context.Context,
	workflowRunID uuid.UUID,
	now time.Time,
) (bool, error) {
	marked := false
	err := h.runRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		run, err := h.runRepo.GetByID(txCtx, workflowRunID)
		if err != nil {
			return err
		}
		if run == nil || run.Status.IsTerminal() {
			return nil
		}

		stepRuns, err := h.stepRunRepo.FindByWorkflowRunID(txCtx, run.ID)
		if err != nil {
			return err
		}

		hasStale := false
		for _, stepRun := range stepRuns {
			if stepRun != nil && stepRun.IsStale(now, h.pendingMaxAge, h.grace) {
				hasStale = true
				break
			}
		}
		if !hasStale {
			return nil
		}

		if err := run.MarkFailed(staleWorkflowRunError); err != nil {
			if errors.Is(err, domainworkflowrun.ErrAlreadyTerminal) {
				return nil
			}
			return err
		}

		events := run.PullEvents()
		for _, stepRun := range stepRuns {
			if stepRun == nil || stepRun.Status.IsTerminal() {
				continue
			}

			if stepRun.IsStale(now, h.pendingMaxAge, h.grace) {
				if err := stepRun.MarkFailed(staleStepRunError, nil, nil); err != nil {
					if errors.Is(err, domainsteprun.ErrAlreadyTerminal) ||
						errors.Is(err, domainsteprun.ErrInvalidStatusTransition) {
						continue
					}
					return err
				}
			} else if err := stepRun.MarkCancelled(); err != nil {
				if errors.Is(err, domainsteprun.ErrAlreadyTerminal) ||
					errors.Is(err, domainsteprun.ErrInvalidStatusTransition) {
					continue
				}
				return err
			}

			if err := h.stepRunRepo.Update(txCtx, stepRun); err != nil {
				return err
			}
			events = append(events, stepRun.PullEvents()...)
		}

		if err := h.runRepo.Update(txCtx, run); err != nil {
			return err
		}
		if err := h.outbox.StoreEvents(txCtx, events); err != nil {
			return err
		}
		marked = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return marked, nil
}
