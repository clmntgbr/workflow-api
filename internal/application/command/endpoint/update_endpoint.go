package endpoint

import (
	"context"
	"errors"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/messaging"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type UpdateEndpointCommand struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         domainendpoint.Method
	Headers        map[string]string
	Query          httpquery.Params
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
	assert *cmdquota.AssertCreateAllowedHandler
}

func NewUpdateEndpointHandler(
	repo domainendpoint.EndpointWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *UpdateEndpointHandler {
	return &UpdateEndpointHandler{repo: repo, outbox: outbox, assert: assert}
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

		if err := h.assert.AssertStepHTTPConfig(txCtx, cmd.UserID, e.ProjectID, cmd.Timeout, cmd.RetryCount); err != nil {
			return err
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
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(e.PullEvents(), cmd.UserID))
	})
}
