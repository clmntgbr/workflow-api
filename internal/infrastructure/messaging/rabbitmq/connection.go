package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	topology Topology
}

func Connect(url string, topology Topology) (*Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}

	if err := topology.declare(ch); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &Connection{conn: conn, channel: ch, topology: topology}, nil
}

func (c *Connection) Channel() *amqp.Channel {
	return c.channel
}

func (c *Connection) Topology() Topology {
	return c.topology
}

func (c *Connection) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
