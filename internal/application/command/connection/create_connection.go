package connection

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type CreateConnectionCommand struct {
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
	SourceStepID   uuid.UUID
	TargetStepID   uuid.UUID
}

type CreateConnectionHandler struct {
	connRepo      domainconnection.ConnectionWriteRepository
	connReadRepo  domainconnection.ConnectionReadRepository
	stepReadRepo  domainstep.StepReadRepository
	stepWriteRepo domainstep.StepWriteRepository
	outbox        port.OutboxRepository
}

func NewCreateConnectionHandler(
	connRepo domainconnection.ConnectionWriteRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	stepReadRepo domainstep.StepReadRepository,
	stepWriteRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *CreateConnectionHandler {
	return &CreateConnectionHandler{
		connRepo:      connRepo,
		connReadRepo:  connReadRepo,
		stepReadRepo:  stepReadRepo,
		stepWriteRepo: stepWriteRepo,
		outbox:        outbox,
	}
}

func (h *CreateConnectionHandler) Handle(
	ctx context.Context,
	cmd CreateConnectionCommand,
) (*domainconnection.Connection, error) {
	if cmd.SourceStepID == cmd.TargetStepID {
		return nil, errors.New("source and target steps must be different")
	}

	source, err := h.stepReadRepo.FindByID(ctx, cmd.SourceStepID)
	if err != nil {
		return nil, errors.New("failed to get source step")
	}
	if source == nil || source.Status == domainstep.StatusDeleted {
		return nil, errors.New("source step not found")
	}
	if source.WorkflowID != cmd.WorkflowID || source.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("source step not found")
	}

	target, err := h.stepReadRepo.FindByID(ctx, cmd.TargetStepID)
	if err != nil {
		return nil, errors.New("failed to get target step")
	}
	if target == nil || target.Status == domainstep.StatusDeleted {
		return nil, errors.New("target step not found")
	}
	if target.WorkflowID != cmd.WorkflowID || target.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("target step not found")
	}

	conn := domainconnection.NewConnection(domainconnection.NewConnectionParams{
		WorkflowID:     cmd.WorkflowID,
		OrganizationID: cmd.OrganizationID,
		SourceStepID:   cmd.SourceStepID,
		TargetStepID:   cmd.TargetStepID,
	})

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list steps")
	}
	connections, err := h.connReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list connections")
	}

	positioned := make([]domainstep.PositionedStep, 0, len(steps))
	for _, stepView := range steps {
		positioned = append(positioned, domainstep.PositionedStep{
			ID:        stepView.ID,
			Position:  stepView.Position,
			CreatedAt: stepView.CreatedAt,
		})
	}
	edges := make([]domainstep.GraphEdge, 0, len(connections)+1)
	for _, connectionView := range connections {
		edges = append(edges, domainstep.GraphEdge{
			SourceStepID: connectionView.SourceStepID,
			TargetStepID: connectionView.TargetStepID,
		})
	}
	edges = append(edges, domainstep.GraphEdge{
		SourceStepID: cmd.SourceStepID,
		TargetStepID: cmd.TargetStepID,
	})
	ordering := domainstep.CalculateOrderingByPosition(positioned, edges)

	executionOrderByStepID := make(map[uuid.UUID]int, len(ordering))
	for stepID, values := range ordering {
		executionOrderByStepID[stepID] = values.ExecutionOrder
	}
	treeIndices := domainstep.CalculateTreeIndices(executionOrderByStepID, edges)

	err = h.connRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.connRepo.Save(txCtx, conn); err != nil {
			return err
		}
		if err := h.stepWriteRepo.UpdateOrdering(txCtx, ordering); err != nil {
			return err
		}
		if err := h.stepWriteRepo.UpdateTreeIndices(txCtx, treeIndices); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, conn.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create connection")
	}

	return conn, nil
}
