package step

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type DeleteStepCommand struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID uuid.UUID
}

type DeleteStepHandler struct {
	stepRepo     domainstep.StepWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connRepo     domainconnection.ConnectionWriteRepository
	connReadRepo domainconnection.ConnectionReadRepository
	variableRepo domainvariable.VariableWriteRepository
	outbox       port.OutboxRepository
}

func NewDeleteStepHandler(
	stepRepo domainstep.StepWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connRepo domainconnection.ConnectionWriteRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	variableRepo domainvariable.VariableWriteRepository,
	outbox port.OutboxRepository,
) *DeleteStepHandler {
	return &DeleteStepHandler{
		stepRepo:     stepRepo,
		stepReadRepo: stepReadRepo,
		connRepo:     connRepo,
		connReadRepo: connReadRepo,
		variableRepo: variableRepo,
		outbox:       outbox,
	}
}

func (h *DeleteStepHandler) Handle(ctx context.Context, cmd DeleteStepCommand) error {
	s, err := h.stepRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get step")
	}
	if s == nil || s.Status == domainstep.StatusDeleted {
		return errors.New("step not found")
	}
	if s.ProjectID != cmd.ProjectID || s.WorkflowID != cmd.WorkflowID {
		return errors.New("step not found")
	}

	connections, err := h.connReadRepo.FindByWorkflowID(ctx, s.WorkflowID)
	if err != nil {
		return errors.New("failed to list connections")
	}

	linked := make([]domainconnection.ConnectionView, 0)
	remainingEdges := make([]domainstep.GraphEdge, 0, len(connections))
	for _, connectionView := range connections {
		if connectionView.SourceStepID == s.ID || connectionView.TargetStepID == s.ID {
			linked = append(linked, connectionView)
			continue
		}
		remainingEdges = append(remainingEdges, domainstep.GraphEdge{
			SourceStepID: connectionView.SourceStepID,
			TargetStepID: connectionView.TargetStepID,
		})
	}

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, s.WorkflowID)
	if err != nil {
		return errors.New("failed to list steps")
	}

	positioned := make([]domainstep.PositionedStep, 0, len(steps))
	for _, stepView := range steps {
		if stepView.ID == s.ID {
			continue
		}
		positioned = append(positioned, domainstep.PositionedStep{
			ID:        stepView.ID,
			Position:  stepView.Position,
			CreatedAt: stepView.CreatedAt,
		})
	}
	ordering := domainstep.CalculateOrderingByPosition(positioned, remainingEdges)
	executionOrderByStepID := make(map[uuid.UUID]int, len(ordering))
	for stepID, values := range ordering {
		executionOrderByStepID[stepID] = values.ExecutionOrder
	}
	treeIndices := domainstep.CalculateTreeIndices(executionOrderByStepID, remainingEdges)

	return h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		events := make([]event.DomainEvent, 0, 1+len(linked))

		for _, connectionView := range linked {
			conn, err := h.connRepo.GetByID(txCtx, connectionView.ID)
			if err != nil {
				return errors.New("failed to get connection")
			}
			if conn == nil {
				continue
			}
			conn.RecordDeletedEvent()
			if err := h.connRepo.Delete(txCtx, conn.ID); err != nil {
				return errors.New("failed to delete connection")
			}
			events = append(events, conn.PullEvents()...)
		}

		s.MarkDeleted()
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return errors.New("failed to delete step")
		}
		events = append(events, s.PullEvents()...)

		if err := h.variableRepo.DeleteByStepID(txCtx, s.ID); err != nil {
			return errors.New("failed to delete variables")
		}

		if err := h.stepRepo.UpdateOrdering(txCtx, ordering); err != nil {
			return err
		}
		if err := h.stepRepo.UpdateTreeIndices(txCtx, treeIndices); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, events)
	})
}
