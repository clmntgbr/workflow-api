package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Topology struct {
	Exchange       string
	Queue          string
	RoutingKey     string
	RetryQueue     string
	DLQ            string
	DLX            string
	RetryTTLMillis int32
}

func DefaultTopology(exchange, queue, routingKey string, retryTTLMillis int) Topology {
	if retryTTLMillis <= 0 {
		retryTTLMillis = 30000
	}
	return Topology{
		Exchange:       exchange,
		Queue:          queue,
		RoutingKey:     routingKey,
		RetryQueue:     queue + ".retry",
		DLQ:            queue + ".dlq",
		DLX:            queue + ".dlx",
		RetryTTLMillis: int32(retryTTLMillis),
	}
}

func (t Topology) declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(t.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", t.Exchange, err)
	}
	if err := ch.ExchangeDeclare(t.DLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx %s: %w", t.DLX, err)
	}

	mainArgs := amqp.Table{
		"x-dead-letter-exchange":    t.DLX,
		"x-dead-letter-routing-key": "retry",
	}
	if _, err := ch.QueueDeclare(t.Queue, true, false, false, false, mainArgs); err != nil {
		return fmt.Errorf("declare queue %s (delete existing queue if args changed): %w", t.Queue, err)
	}
	if err := ch.QueueBind(t.Queue, t.RoutingKey, t.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", t.Queue, err)
	}
	if err := ch.QueueBind(t.Queue, "main", t.DLX, false, nil); err != nil {
		return fmt.Errorf("bind queue %s to dlx main: %w", t.Queue, err)
	}

	retryArgs := amqp.Table{
		"x-message-ttl":             t.RetryTTLMillis,
		"x-dead-letter-exchange":    t.DLX,
		"x-dead-letter-routing-key": "main",
	}
	if _, err := ch.QueueDeclare(t.RetryQueue, true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("declare retry queue %s: %w", t.RetryQueue, err)
	}
	if err := ch.QueueBind(t.RetryQueue, "retry", t.DLX, false, nil); err != nil {
		return fmt.Errorf("bind retry queue: %w", err)
	}

	if _, err := ch.QueueDeclare(t.DLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq %s: %w", t.DLQ, err)
	}
	if err := ch.QueueBind(t.DLQ, "dlq", t.DLX, false, nil); err != nil {
		return fmt.Errorf("bind dlq: %w", err)
	}

	return nil
}
