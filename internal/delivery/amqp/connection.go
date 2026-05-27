package amqp

import (
	"context"
	"fmt"

	"go-arch/internal/config"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	conn *rabbitmq.Connection
}

func NewConnection(url string) (*Connection, error) {
	conn, err := rabbitmq.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connect rabbitmq: %w", err)
	}

	return &Connection{conn: conn}, nil
}

func (c *Connection) Close() error {
	return c.conn.Close()
}

func (c *Connection) Channel() (*rabbitmq.Channel, error) {
	return c.conn.Channel()
}

func (c *Connection) Name() string {
	return "rabbitmq"
}

func (c *Connection) Ping(ctx context.Context) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func DeclareTopology(ctx context.Context, conn *Connection, cfg config.RabbitMQConfig) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(
		cfg.Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	queue, err := ch.QueueDeclare(
		cfg.UserCreatedQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(
		queue.Name,
		cfg.UserCreatedRoutingKey,
		cfg.Exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("bind queue: %w", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
