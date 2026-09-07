package workflow

import (
	"context"
	"time"

	"go-api/internal/domain/port"
	domainworkflow "go-api/internal/domain/workflow"
)

type ClaimDueWorkflowsCommand struct {
	Now   time.Time
	Limit int
}

type ClaimDueWorkflowsHandler struct {
	repo   domainworkflow.WorkflowWriteRepository
	outbox port.OutboxRepository
}

func NewClaimDueWorkflowsHandler(
	repo domainworkflow.WorkflowWriteRepository,
	outbox port.OutboxRepository,
) *ClaimDueWorkflowsHandler {
	return &ClaimDueWorkflowsHandler{repo: repo, outbox: outbox}
}

func (h *ClaimDueWorkflowsHandler) Handle(
	ctx context.Context,
	cmd ClaimDueWorkflowsCommand,
) ([]*domainworkflow.Workflow, error) {
	var claimed []*domainworkflow.Workflow
	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		var err error
		claimed, err = h.repo.ClaimDueForExecution(txCtx, cmd.Now, cmd.Limit)
		if err != nil {
			return err
		}
		for _, w := range claimed {
			if err := h.outbox.StoreEvents(txCtx, w.PullEvents()); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}
