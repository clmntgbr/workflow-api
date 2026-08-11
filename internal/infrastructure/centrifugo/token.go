package centrifugo

import (
	"fmt"
	"time"

	"go-api/internal/infrastructure/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ConnectionInfo struct {
	Token   string `json:"token"`
	Channel string `json:"channel"`
	WSURL   string `json:"wsUrl"`
}

func NewConnectionInfo(env *config.Config, userID uuid.UUID) (ConnectionInfo, error) {
	channel := UserChannel(userID)
	token, err := generateConnectionToken(env.CentrifugoTokenSecret, userID, channel)
	if err != nil {
		return ConnectionInfo{}, err
	}

	return ConnectionInfo{
		Token:   token,
		Channel: channel,
		WSURL:   env.CentrifugoPublicWSURL,
	}, nil
}

func generateConnectionToken(secret string, userID uuid.UUID, channel string) (string, error) {
	claims := jwt.MapClaims{
		"sub":      userID.String(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"channels": []string{channel},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign centrifugo token: %w", err)
	}

	return signed, nil
}
