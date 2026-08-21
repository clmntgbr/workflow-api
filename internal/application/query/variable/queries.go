package variable

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type GetVariableByIDQuery struct {
	ID         uuid.UUID
	WorkflowID uuid.UUID
}

type GetVariableByIDHandler struct {
	readRepo domainvariable.VariableReadRepository
}

func NewGetVariableByIDHandler(readRepo domainvariable.VariableReadRepository) *GetVariableByIDHandler {
	return &GetVariableByIDHandler{readRepo: readRepo}
}

func (h *GetVariableByIDHandler) Handle(ctx context.Context, q GetVariableByIDQuery) (*domainvariable.VariableView, error) {
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get variable")
	}
	if view == nil || view.WorkflowID != q.WorkflowID {
		return nil, domainvariable.ErrNotFound
	}
	return view, nil
}

type ListVariablesByWorkflowQuery struct {
	WorkflowID uuid.UUID
}

type ListVariablesByWorkflowHandler struct {
	readRepo domainvariable.VariableReadRepository
}

func NewListVariablesByWorkflowHandler(readRepo domainvariable.VariableReadRepository) *ListVariablesByWorkflowHandler {
	return &ListVariablesByWorkflowHandler{readRepo: readRepo}
}

func (h *ListVariablesByWorkflowHandler) Handle(
	ctx context.Context,
	q ListVariablesByWorkflowQuery,
) ([]domainvariable.VariableView, error) {
	views, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list variables")
	}
	return views, nil
}

type ListAvailableVariablesQuery struct {
	WorkflowID uuid.UUID
	StepID     uuid.UUID
}

type ListAvailableVariablesHandler struct {
	readRepo     domainvariable.VariableReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
}

func NewListAvailableVariablesHandler(
	readRepo domainvariable.VariableReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
) *ListAvailableVariablesHandler {
	return &ListAvailableVariablesHandler{readRepo: readRepo, connReadRepo: connReadRepo}
}

func (h *ListAvailableVariablesHandler) Handle(
	ctx context.Context,
	q ListAvailableVariablesQuery,
) ([]domainvariable.VariableView, error) {
	connections, err := h.connReadRepo.FindByWorkflowID(ctx, q.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list connections")
	}
	edges := make([]domainvariable.GraphEdge, 0, len(connections))
	for _, conn := range connections {
		edges = append(edges, domainvariable.GraphEdge{
			SourceStepID: conn.SourceStepID,
			TargetStepID: conn.TargetStepID,
		})
	}
	ancestors := domainvariable.AncestorStepIDs(q.StepID, edges)
	all, err := h.readRepo.FindByWorkflowID(ctx, q.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list variables")
	}
	out := make([]domainvariable.VariableView, 0)
	for _, view := range all {
		if view.Kind == domainvariable.KindStatic {
			out = append(out, view)
			continue
		}
		if view.StepID == nil {
			continue
		}
		if _, ok := ancestors[*view.StepID]; ok {
			out = append(out, view)
		}
	}
	return out, nil
}
