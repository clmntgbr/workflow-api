package endpoint

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/event"
	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

const (
	maxEndpointName        = 255
	maxEndpointDescription = 2000
	maxEndpointURL         = 2048
	maxImportOperations    = 1000
	maxNameCollisionTries  = 50
)

type ImportEndpointsFromOpenAPICommand struct {
	Spec           []byte
	BaseURL        string
	Status         domainendpoint.Status
	Headers        map[string]string
	Query          httpquery.Params
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	OrganizationID uuid.UUID
}

type ImportEndpointsFromOpenAPIHandler struct {
	repo   domainendpoint.EndpointWriteRepository
	parser port.OpenAPISpecParser
	outbox port.OutboxRepository
}

func NewImportEndpointsFromOpenAPIHandler(
	repo domainendpoint.EndpointWriteRepository,
	parser port.OpenAPISpecParser,
	outbox port.OutboxRepository,
) *ImportEndpointsFromOpenAPIHandler {
	return &ImportEndpointsFromOpenAPIHandler{
		repo:   repo,
		parser: parser,
		outbox: outbox,
	}
}

func (h *ImportEndpointsFromOpenAPIHandler) Handle(
	ctx context.Context,
	cmd ImportEndpointsFromOpenAPICommand,
) ([]domainendpoint.Endpoint, error) {
	if cmd.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}
	if strings.TrimSpace(cmd.BaseURL) == "" {
		return nil, errors.New("baseUrl is required")
	}
	if cmd.Status != domainendpoint.StatusActive && cmd.Status != domainendpoint.StatusInactive {
		return nil, errors.New("invalid status")
	}

	operations, err := h.parser.Parse(ctx, cmd.Spec)
	if err != nil {
		return nil, err
	}
	if len(operations) > maxImportOperations {
		return nil, domainendpoint.ErrTooManyOperations
	}

	usedNames := map[string]struct{}{}
	prepared := make([]domainendpoint.NewEndpointParams, 0, len(operations))
	for _, operation := range operations {
		method, err := domainendpoint.ParseMethod(operation.Method)
		if err != nil {
			continue
		}

		endpointURL := domainendpoint.JoinBaseURLAndPath(cmd.BaseURL, operation.Path)
		if endpointURL == "" || utf8.RuneCountInString(endpointURL) > maxEndpointURL {
			return nil, fmt.Errorf("%w: path %s", domainendpoint.ErrInvalidEndpointURL, operation.Path)
		}

		prepared = append(prepared, domainendpoint.NewEndpointParams{
			Name:             uniqueEndpointName(usedNames, operation),
			Description:      truncateRunes(firstNonEmpty(operation.Description, operation.Summary), maxEndpointDescription),
			URL:              endpointURL,
			Method:           method,
			Headers:          cloneStringMap(cmd.Headers),
			Query:            httpquery.Clone(cmd.Query),
			Body:             cloneAnyMap(cmd.Body),
			Timeout:          cmd.Timeout,
			RetryOnFailure:   cmd.RetryOnFailure,
			RetryCount:       cmd.RetryCount,
			RetryDelay:       cmd.RetryDelay,
			Status:           cmd.Status,
			SkipCreatedEvent: true,
			OrganizationID:   cmd.OrganizationID,
		})
	}
	if len(prepared) == 0 {
		return nil, domainendpoint.ErrNoOperations
	}

	created := make([]domainendpoint.Endpoint, 0, len(prepared))
	err = h.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		var events []event.DomainEvent
		for _, params := range prepared {
			e := domainendpoint.NewEndpoint(params)
			if err := h.repo.Save(txCtx, e); err != nil {
				return err
			}
			_ = e.PullEvents()
			created = append(created, *e)
		}
		events = append(events, domainendpoint.EndpointImported{
			ID:             uuid.NewString(),
			OrganizationID: cmd.OrganizationID.String(),
			Count:          len(created),
			Timestamp:      time.Now().UTC(),
		})
		return h.outbox.StoreEvents(txCtx, events)
	})
	if err != nil {
		return nil, errors.New("failed to import endpoints")
	}

	return created, nil
}

func uniqueEndpointName(used map[string]struct{}, operation port.OpenAPIOperation) string {
	base := firstNonEmpty(operation.Summary, operation.OperationID, operation.Method+" "+operation.Path)
	candidates := []string{
		base,
		base + " (" + operation.Method + ")",
		base + " (" + operation.Method + " " + operation.Path + ")",
	}
	for _, candidate := range candidates {
		if name := truncateRunes(candidate, maxEndpointName); reserveName(used, name) {
			return name
		}
	}
	for i := 2; i <= maxNameCollisionTries; i++ {
		suffix := fmt.Sprintf(" (%d)", i)
		limit := maxEndpointName - utf8.RuneCountInString(suffix)
		name := truncateRunes(base, limit) + suffix
		if reserveName(used, name) {
			return name
		}
	}
	return truncateRunes(operation.Method+" "+operation.Path+" "+uuid.NewString(), maxEndpointName)
}

func reserveName(used map[string]struct{}, name string) bool {
	if name == "" {
		return false
	}
	if _, exists := used[name]; exists {
		return false
	}
	used[name] = struct{}{}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func truncateRunes(value string, max int) string {
	if max <= 0 || utf8.RuneCountInString(value) <= max {
		return value
	}
	return string([]rune(value)[:max])
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
