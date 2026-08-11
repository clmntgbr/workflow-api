package notification

import (
	"context"
	"log"

	"go-api/internal/domain/port"
)

type LogNotifier struct{}

func NewLogNotifier() port.NotificationSender {
	return &LogNotifier{}
}

func (n *LogNotifier) Send(ctx context.Context, userID, message string) error {
	log.Printf("notification userId=%s message=%s", userID, message)
	return nil
}
