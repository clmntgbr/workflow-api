package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"

	amqp "github.com/rabbitmq/amqp091-go"
)

const defaultMaxRetries = 3

type Consumer struct {
	channel     *amqp.Channel
	topology    Topology
	registry    *registry.HandlerRegistry
	concurrency int
	maxRetries  int
}

func NewConsumer(conn *Connection, reg *registry.HandlerRegistry, concurrency, maxRetries int) *Consumer {
	if concurrency <= 0 {
		concurrency = 1
	}
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	return &Consumer{
		channel:     conn.Channel(),
		topology:    conn.Topology(),
		registry:    reg,
		concurrency: concurrency,
		maxRetries:  maxRetries,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	if err := c.channel.Qos(c.concurrency, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	msgs, err := c.channel.Consume(c.topology.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Printf(
		"rabbitmq consumer started on queue %s (concurrency=%d maxRetries=%d dlq=%s)",
		c.topology.Queue,
		c.concurrency,
		c.maxRetries,
		c.topology.DLQ,
	)

	sem := make(chan struct{}, c.concurrency)
	for {
		select {
		case <-ctx.Done():
			log.Println("rabbitmq consumer stopped")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			sem <- struct{}{}
			go func(m amqp.Delivery) {
				defer func() { <-sem }()
				c.handleDelivery(ctx, m)
			}(msg)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, m amqp.Delivery) {
	var envelope port.EventEnvelope
	if err := json.Unmarshal(m.Body, &envelope); err != nil {
		log.Printf("invalid event envelope: %v", err)
		c.sendToDLQ(m, "invalid_envelope")
		_ = m.Ack(false)
		return
	}

	err := c.registry.Dispatch(ctx, envelope.Type, envelope.Payload)
	if err == nil {
		_ = m.Ack(false)
		return
	}

	var nonRetryable *messaging.NonRetryableError
	attempts := retryCount(m)
	if errors.As(err, &nonRetryable) || attempts >= c.maxRetries {
		log.Printf(
			"moving event to DLQ type=%s eventId=%s attempts=%d err=%v",
			envelope.Type,
			envelope.EventID,
			attempts,
			err,
		)
		c.sendToDLQ(m, "handler_failed")
		_ = m.Ack(false)
		return
	}

	log.Printf(
		"retryable failure type=%s eventId=%s attempts=%d err=%v → retry queue",
		envelope.Type,
		envelope.EventID,
		attempts,
		err,
	)
	// Never requeue=true: dead-letter to retry queue (TTL), then back to main.
	_ = m.Nack(false, false)
}

func (c *Consumer) sendToDLQ(m amqp.Delivery, reason string) {
	headers := amqp.Table{}
	for k, v := range m.Headers {
		headers[k] = v
	}
	headers["x-dlq-reason"] = reason

	err := c.channel.PublishWithContext(
		context.Background(),
		c.topology.DLX,
		"dlq",
		false,
		false,
		amqp.Publishing{
			ContentType:  m.ContentType,
			DeliveryMode: amqp.Persistent,
			Type:         m.Type,
			Headers:      headers,
			Body:         m.Body,
		},
	)
	if err != nil {
		log.Printf("failed to publish to DLQ: %v", err)
	}
}

func retryCount(msg amqp.Delivery) int {
	death, ok := msg.Headers["x-death"].([]interface{})
	if !ok || len(death) == 0 {
		return 0
	}
	if entry, ok := death[0].(amqp.Table); ok {
		switch count := entry["count"].(type) {
		case int64:
			return int(count)
		case int32:
			return int(count)
		case int:
			return count
		case float64:
			return int(count)
		}
	}
	return 0
}
