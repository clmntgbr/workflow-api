package step

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type UpdateStepPositionCommand struct {
	ID         uuid.UUID
	ProjectID  uuid.UUID
	WorkflowID uuid.UUID
	Position   domainstep.Position
}

type UpdateStepPositionHandler struct {
	stepRepo     domainstep.StepWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
	outbox       port.OutboxRepository
}

func NewUpdateStepPositionHandler(
	stepRepo domainstep.StepWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	outbox port.OutboxRepository,
) *UpdateStepPositionHandler {
	return &UpdateStepPositionHandler{
		stepRepo:     stepRepo,
		stepReadRepo: stepReadRepo,
		connReadRepo: connReadRepo,
		outbox:       outbox,
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
	if s.ProjectID != cmd.ProjectID || s.WorkflowID != cmd.WorkflowID {
		return nil, errors.New("step not found")
	}

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, s.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list steps")
	}
	connections, err := h.connReadRepo.FindByWorkflowID(ctx, s.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list connections")
	}

	positioned := make([]domainstep.PositionedStep, 0, len(steps))
	for _, stepView := range steps {
		position := stepView.Position
		if stepView.ID == s.ID {
			position = cmd.Position
		}
		positioned = append(positioned, domainstep.PositionedStep{
			ID:        stepView.ID,
			Position:  position,
			CreatedAt: stepView.CreatedAt,
		})
	}
	edges := make([]domainstep.GraphEdge, 0, len(connections))
	for _, connectionView := range connections {
		edges = append(edges, domainstep.GraphEdge{
			SourceStepID: connectionView.SourceStepID,
			TargetStepID: connectionView.TargetStepID,
		})
	}
	ordering := domainstep.CalculateOrderingByPosition(positioned, edges)

	executionOrderByStepID := make(map[uuid.UUID]int, len(ordering))
	for stepID, values := range ordering {
		executionOrderByStepID[stepID] = values.ExecutionOrder
	}
	treeIndices := domainstep.CalculateTreeIndices(executionOrderByStepID, edges)

	s.ApplyPositionUpdate(ordering[s.ID].Index, cmd.Position)
	s.TreeIndex = treeIndices[s.ID]

	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return err
		}
		if err := h.stepRepo.UpdateOrdering(txCtx, ordering); err != nil {
			return err
		}
		if err := h.stepRepo.UpdateTreeIndices(txCtx, treeIndices); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, s.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	return s, nil
}
