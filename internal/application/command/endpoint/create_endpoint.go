package endpoint

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type CreateEndpointCommand struct {
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
	OrganizationID uuid.UUID
}

type CreateEndpointHandler struct {
	repo   domainendpoint.EndpointWriteRepository
	outbox port.OutboxRepository
}

func NewCreateEndpointHandler(
	repo domainendpoint.EndpointWriteRepository,
	outbox port.OutboxRepository,
) *CreateEndpointHandler {
	return &CreateEndpointHandler{repo: repo, outbox: outbox}
}

func (h *CreateEndpointHandler) Handle(
	ctx context.Context,
	cmd CreateEndpointCommand,
) (*domainendpoint.Endpoint, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.URL == "" {
		return nil, errors.New("url is required")
	}
	if !cmd.Method.Valid() {
		return nil, errors.New("invalid method")
	}
	if cmd.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}

	e := domainendpoint.NewEndpoint(domainendpoint.NewEndpointParams{
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
		OrganizationID: cmd.OrganizationID,
	})

	err := h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.repo.Save(txCtx, e); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, e.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create endpoint")
	}

	return e, nil
}
