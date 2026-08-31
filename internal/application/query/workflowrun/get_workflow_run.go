package workflowrun

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type GetWorkflowRunByIDQuery struct {
	UserID    uuid.UUID
	ProjectID uuid.UUID
	ID        uuid.UUID
}

type GetWorkflowRunByIDHandler struct {
	readRepo  domainworkflowrun.WorkflowRunReadRepository
	retention querysubscription.RunHistoryCutoffResolver
}

func NewGetWorkflowRunByIDHandler(
	readRepo domainworkflowrun.WorkflowRunReadRepository,
	retention querysubscription.RunHistoryCutoffResolver,
) *GetWorkflowRunByIDHandler {
	return &GetWorkflowRunByIDHandler{readRepo: readRepo, retention: retention}
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

	cutoff, err := h.retention.RunHistoryCutoff(ctx, q.UserID, q.ProjectID)
	if err != nil {
		return nil, errors.New("failed to resolve run history retention")
	}
	if !querysubscription.IsWithinRunHistory(view.CreatedAt, cutoff) {
		return nil, errors.New("workflow run not found")
	}

	return view, nil
}
