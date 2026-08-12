package endpoint

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type DeleteEndpointCommand struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
}

type DeleteEndpointHandler struct {
	repo   domainendpoint.EndpointWriteRepository
	outbox port.OutboxRepository
}

func NewDeleteEndpointHandler(
	repo domainendpoint.EndpointWriteRepository,
	outbox port.OutboxRepository,
) *DeleteEndpointHandler {
	return &DeleteEndpointHandler{repo: repo, outbox: outbox}
}

func (h *DeleteEndpointHandler) Handle(ctx context.Context, cmd DeleteEndpointCommand) error {
	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		e, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get endpoint")
		}
		if e == nil || e.Status == domainendpoint.StatusDeleted {
			return errors.New("endpoint not found")
		}
		if e.OrganizationID != cmd.OrganizationID {
			return errors.New("endpoint not found")
		}

		e.MarkDeleted()

		if err := h.repo.Update(txCtx, e); err != nil {
			return errors.New("failed to delete endpoint")
		}
		return h.outbox.StoreEvents(txCtx, e.PullEvents())
	})
}
