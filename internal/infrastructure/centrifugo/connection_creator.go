package centrifugo

import (
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"

	"github.com/google/uuid"
)

type ConnectionInfoCreator struct {
	env *config.Config
}

func NewConnectionInfoCreator(env *config.Config) *ConnectionInfoCreator {
	return &ConnectionInfoCreator{env: env}
}

func (c *ConnectionInfoCreator) CreateConnectionInfo(userID uuid.UUID) (port.RealtimeConnection, error) {
	info, err := NewConnectionInfo(c.env, userID)
	if err != nil {
		return port.RealtimeConnection{}, err
	}

	return port.RealtimeConnection{
		Token:   info.Token,
		Channel: info.Channel,
		WSURL:   info.WSURL,
	}, nil
}
