package handler

import (
	"go-api/internal/infrastructure/centrifugo"

	"github.com/google/uuid"
)

type realtimeConnectionCreator interface {
	CreateConnectionInfo(userID uuid.UUID) (centrifugo.ConnectionInfo, error)
}
