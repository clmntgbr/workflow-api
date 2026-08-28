package steprun

import (
	"context"
	"time"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"
)

type ResumeDueWaitingStepRunsHandler struct {
	repo   domainsteprun.StepRunWriteRepository
	outbox port.OutboxRepository
}

func NewResumeDueWaitingStepRunsHandler(
	repo domainsteprun.StepRunWriteRepository,
	outbox port.OutboxRepository,
) *ResumeDueWaitingStepRunsHandler {
	return &ResumeDueWaitingStepRunsHandler{
		repo:   repo,
		outbox: outbox,
	}
}

func (h *ResumeDueWaitingStepRunsHandler) Handle(ctx context.Context, now time.Time, limit int) (int, error) {
	claimed, err := h.repo.ClaimDueWaiting(ctx, now, limit)
	if err != nil {
		return 0, err
	}

	resumed := 0
	for _, run := range claimed {
		if run == nil {
			continue
		}

		processed := false
		err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
			current, err := h.repo.GetByID(txCtx, run.ID)
			if err != nil {
				return err
			}
			if current == nil ||
				current.Status != domainsteprun.StatusWaiting ||
				current.ResumeAt == nil ||
				current.ResumeAt.After(now.UTC()) {
				return nil
			}

			if err := current.MarkStarted(); err != nil {
				return err
			}
			if err := current.MarkSucceeded(
				domainsteprun.ResponseSnapshot{},
				map[string]any{},
				[]domainassertion.Result{},
			); err != nil {
				return err
			}
			if err := h.repo.Update(txCtx, current); err != nil {
				return err
			}
			processed = true
			return h.outbox.StoreEvents(txCtx, current.PullEvents())
		})
		if err != nil {
			return resumed, err
		}
		if processed {
			resumed++
		}
	}
	return resumed, nil
}
