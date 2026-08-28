package connection

import (
	"context"
	"errors"

	"go-api/internal/application/messaging"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type DeleteConnectionHandler struct {
	connRepo      domainconnection.ConnectionWriteRepository
	connReadRepo  domainconnection.ConnectionReadRepository
	stepReadRepo  domainstep.StepReadRepository
	stepWriteRepo domainstep.StepWriteRepository
	outbox        port.OutboxRepository
}

func NewDeleteConnectionHandler(
	connRepo domainconnection.ConnectionWriteRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	stepReadRepo domainstep.StepReadRepository,
	stepWriteRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *DeleteConnectionHandler {
	return &DeleteConnectionHandler{
		connRepo:      connRepo,
		connReadRepo:  connReadRepo,
		stepReadRepo:  stepReadRepo,
		stepWriteRepo: stepWriteRepo,
		outbox:        outbox,
	}
}

type DeleteConnectionCommand struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	WorkflowID uuid.UUID
	ProjectID uuid.UUID
}

func (h *DeleteConnectionHandler) Handle(ctx context.Context, cmd DeleteConnectionCommand) error {
	conn, err := h.connRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return errors.New("failed to get connection")
	}
	if conn == nil {
		return errors.New("connection not found")
	}
	if conn.ProjectID != cmd.ProjectID || conn.WorkflowID != cmd.WorkflowID {
		return errors.New("connection not found")
	}

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, conn.WorkflowID)
	if err != nil {
		return errors.New("failed to list steps")
	}
	connections, err := h.connReadRepo.FindByWorkflowID(ctx, conn.WorkflowID)
	if err != nil {
		return errors.New("failed to list connections")
	}

	positioned := make([]domainstep.PositionedStep, 0, len(steps))
	for _, stepView := range steps {
		positioned = append(positioned, domainstep.PositionedStep{
			ID:        stepView.ID,
			Position:  stepView.Position,
			CreatedAt: stepView.CreatedAt,
		})
	}
	edges := make([]domainstep.GraphEdge, 0, len(connections))
	for _, connectionView := range connections {
		if connectionView.ID == cmd.ID {
			continue
		}
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

	return h.connRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		conn.RecordDeletedEvent()

		if err := h.connRepo.Delete(txCtx, conn.ID); err != nil {
			return errors.New("failed to delete connection")
		}
		if err := h.stepWriteRepo.UpdateOrdering(txCtx, ordering); err != nil {
			return err
		}
		if err := h.stepWriteRepo.UpdateTreeIndices(txCtx, treeIndices); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(conn.PullEvents(), cmd.UserID))
	})
}
