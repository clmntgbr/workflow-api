package port

import "context"

type NotificationSender interface {
	Send(ctx context.Context, userID, message string) error
}
