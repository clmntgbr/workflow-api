package presenter

import (
	"go-api/internal/domain/port"
)

type RealtimeConnectionResponse struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
	WSURL   string `json:"wsUrl"`
}

func NewRealtimeConnectionResponse(connection port.RealtimeConnection) RealtimeConnectionResponse {
	return RealtimeConnectionResponse{
		Token:   connection.Token,
		Channel: connection.Channel,
		WSURL:   connection.WSURL,
	}
}
