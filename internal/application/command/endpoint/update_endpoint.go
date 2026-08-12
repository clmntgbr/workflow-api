package endpoint

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type UpdateEndpointCommand struct {
	ID             uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         domainendpoint.Method
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         domainendpoint.Status
}

type UpdateEndpointHandler struct {
	repo   domainendpoint.EndpointWriteRepository
	outbox port.OutboxRepository
}

func NewUpdateEndpointHandler(
	repo domainendpoint.EndpointWriteRepository,
	outbox port.OutboxRepository,
) *UpdateEndpointHandler {
	return &UpdateEndpointHandler{repo: repo, outbox: outbox}
}

func (h *UpdateEndpointHandler) Handle(ctx context.Context, cmd UpdateEndpointCommand) error {
	if cmd.Name == "" {
		return errors.New("name is required")
	}
	if cmd.URL == "" {
		return errors.New("url is required")
	}
	if !cmd.Method.Valid() {
		return errors.New("invalid method")
	}
	if !cmd.Status.Valid() {
		return errors.New("invalid status")
	}
	if cmd.Status == domainendpoint.StatusDeleted {
		return errors.New("use delete to mark an endpoint as deleted")
	}

	return h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		e, err := h.repo.GetByID(txCtx, cmd.ID)
		if err != nil {
			return errors.New("failed to get endpoint")
		}
		if e == nil || e.Status == domainendpoint.StatusDeleted {
			return errors.New("endpoint not found")
		}

		e.ApplyUpdate(domainendpoint.UpdateEndpointParams{
			Name:           cmd.Name,
			Description:    cmd.Description,
			URL:            cmd.URL,
			Method:         cmd.Method,
			Headers:        cmd.Headers,
			Query:          cmd.Query,
			Body:           cmd.Body,
			Timeout:        cmd.Timeout,
			RetryOnFailure: cmd.RetryOnFailure,
			RetryCount:     cmd.RetryCount,
			RetryDelay:     cmd.RetryDelay,
			Status:         cmd.Status,
		})

		if err := h.repo.Update(txCtx, e); err != nil {
			return errors.New("failed to update endpoint")
		}
		return h.outbox.StoreEvents(txCtx, e.PullEvents())
	})
}
