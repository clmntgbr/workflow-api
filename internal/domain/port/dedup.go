package port

import "context"

// ProcessedEventRepository deduplicates handler side effects per (event_id, handler_name).
type ProcessedEventRepository interface {
	// MarkProcessed inserts a dedup row. inserted=false means already processed.
	MarkProcessed(ctx context.Context, eventID, handlerName string) (inserted bool, err error)
}
