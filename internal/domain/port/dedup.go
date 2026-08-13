package port

import "context"

type ProcessedEventRepository interface {
	MarkProcessed(ctx context.Context, eventID, handlerName string) (inserted bool, err error)
}
