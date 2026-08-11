package centrifugo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go-api/internal/infrastructure/config"

	"github.com/centrifugal/gocent/v3"
	"github.com/google/uuid"
)

type Publisher struct {
	client *gocent.Client
}

func NewPublisher(env *config.Config) *Publisher {
	return &Publisher{
		client: gocent.New(gocent.Config{
			Addr: apiEndpoint(env.CentrifugoURL),
			Key:  env.CentrifugoAPIKey,
		}),
	}
}

func apiEndpoint(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/api") {
		return baseURL
	}
	return baseURL + "/api"
}

func (p *Publisher) PublishToUser(ctx context.Context, userID uuid.UUID, eventType string, payload any) error {
	body, err := marshalRealtimeEvent(eventType, payload)
	if err != nil {
		return fmt.Errorf("failed to marshal centrifugo event: %w", err)
	}

	channel := UserChannel(userID)
	if _, err := p.client.Publish(ctx, channel, body); err != nil {
		return fmt.Errorf("failed to publish to centrifugo channel %q: %w", channel, err)
	}
	return nil
}

func marshalRealtimeEvent(eventType string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	fields["type"] = eventType
	return json.Marshal(fields)
}
