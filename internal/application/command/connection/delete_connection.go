package connection

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type DeleteConnectionHandler struct {
	connRepo domainconnection.ConnectionWriteRepository
	outbox   port.OutboxRepository
}

func NewDeleteConnectionHandler(
	connRepo domainconnection.ConnectionWriteRepository,
	outbox port.OutboxRepository,
) *DeleteConnectionHandler {
	return &DeleteConnectionHandler{connRepo: connRepo, outbox: outbox}
}

func (h *DeleteConnectionHandler) Handle(ctx context.Context, id uuid.UUID) error {
	return h.connRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		conn, err := h.connRepo.GetByID(txCtx, id)
		if err != nil {
			return errors.New("failed to get connection")
		}
		if conn == nil {
			return nil
		}

		conn.RecordDeletedEvent()

		if err := h.connRepo.Delete(txCtx, conn.ID); err != nil {
			return errors.New("failed to delete connection")
		}
		return h.outbox.StoreEvents(txCtx, conn.PullEvents())
	})
}
