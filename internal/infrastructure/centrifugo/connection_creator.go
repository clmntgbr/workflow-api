package centrifugo

import (
	"go-api/internal/infrastructure/config"

	"github.com/google/uuid"
)

type ConnectionInfoCreator struct {
	env *config.Config
}

func NewConnectionInfoCreator(env *config.Config) *ConnectionInfoCreator {
	return &ConnectionInfoCreator{env: env}
}

func (c *ConnectionInfoCreator) CreateConnectionInfo(userID uuid.UUID) (ConnectionInfo, error) {
	return NewConnectionInfo(c.env, userID)
}
