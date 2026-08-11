package dedup

import (
	"context"
	"encoding/json"
	"errors"

	"go-api/internal/application/messaging"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"
)

type eventIDPayload struct {
	EventID string `json:"eventId"`
}

type processedEventReleaser interface {
	port.ProcessedEventRepository
	UnmarkProcessed(ctx context.Context, eventID, handlerName string) error
}

// With wraps a handler so each (eventId, handlerName) is processed at most once.
// Mark happens before next; on retryable failure the mark is released so the event can retry.
func With(repo port.ProcessedEventRepository, handlerName string, next registry.EventHandler) registry.EventHandler {
	return func(ctx context.Context, payload []byte) error {
		var meta eventIDPayload
		if err := json.Unmarshal(payload, &meta); err != nil {
			return messaging.NonRetryable(err)
		}
		if meta.EventID == "" {
			return messaging.NonRetryable(errMissingEventID)
		}

		inserted, err := repo.MarkProcessed(ctx, meta.EventID, handlerName)
		if err != nil {
			return messaging.Retryable(err)
		}
		if !inserted {
			return nil
		}

		if err := next(ctx, payload); err != nil {
			var retryable *messaging.RetryableError
			if errors.As(err, &retryable) {
				if releaser, ok := repo.(processedEventReleaser); ok {
					_ = releaser.UnmarkProcessed(ctx, meta.EventID, handlerName)
				}
			}
			return err
		}
		return nil
	}
}

type missingEventIDError struct{}

func (missingEventIDError) Error() string { return "missing eventId in payload" }

var errMissingEventID error = missingEventIDError{}
