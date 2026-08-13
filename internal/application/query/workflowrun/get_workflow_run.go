package workflowrun

import (
	"context"
	"errors"

	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type GetWorkflowRunByIDQuery struct {
	ID uuid.UUID
}

type GetWorkflowRunByIDHandler struct {
	readRepo domainworkflowrun.WorkflowRunReadRepository
}

func NewGetWorkflowRunByIDHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
) *GetWorkflowRunByIDHandler {
	return &GetWorkflowRunByIDHandler{readRepo: readRepo}
}

func (h *GetWorkflowRunByIDHandler) Handle(
	ctx context.Context,
	q GetWorkflowRunByIDQuery,
) (*domainworkflowrun.WorkflowRunView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get workflow run")
	}
	if view == nil {
		return nil, errors.New("workflow run not found")
	}
	return view, nil
}
