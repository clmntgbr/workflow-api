package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"

	"go-api/internal/domain/port"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	channel  *amqp.Channel
	exchange string
}

func NewPublisher(conn *Connection, exchange string) *Publisher {
	return &Publisher{channel: conn.Channel(), exchange: exchange}
}

func (p *Publisher) Publish(ctx context.Context, envelope port.EventEnvelope) error {
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	routingKey := envelope.Type
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Type:         envelope.Type,
		Body:         body,
	}
	if envelope.Priority > 0 {
		publishing.Priority = envelope.Priority
	}
	err = p.channel.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		publishing,
	)
	if err != nil {
		return fmt.Errorf("publish %s: %w", envelope.Type, err)
	}
	return nil
}
