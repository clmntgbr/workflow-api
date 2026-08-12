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
	connRepo domainconnection.ConnectionWriteRepository
	stepRepo domainstep.StepReadRepository
	outbox   port.OutboxRepository
}

func NewCreateConnectionHandler(
	connRepo domainconnection.ConnectionWriteRepository,
	stepRepo domainstep.StepReadRepository,
	outbox port.OutboxRepository,
) *CreateConnectionHandler {
	return &CreateConnectionHandler{
		connRepo: connRepo,
		stepRepo: stepRepo,
		outbox:   outbox,
	}
}

func (h *CreateConnectionHandler) Handle(
	ctx context.Context,
	cmd CreateConnectionCommand,
) (*domainconnection.Connection, error) {
	if cmd.SourceStepID == cmd.TargetStepID {
		return nil, errors.New("source and target steps must be different")
	}

	source, err := h.stepRepo.FindByID(ctx, cmd.SourceStepID)
	if err != nil {
		return nil, errors.New("failed to get source step")
	}
	if source == nil || source.Status == domainstep.StatusDeleted {
		return nil, errors.New("source step not found")
	}
	if source.WorkflowID != cmd.WorkflowID || source.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("source step not found")
	}

	target, err := h.stepRepo.FindByID(ctx, cmd.TargetStepID)
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

	err = h.connRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.connRepo.Save(txCtx, conn); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, conn.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create connection")
	}

	return conn, nil
}
