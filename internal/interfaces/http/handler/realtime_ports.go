package handler

import (
	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type realtimeConnectionCreator interface {
	CreateConnectionInfo(userID uuid.UUID) (port.RealtimeConnection, error)
}
