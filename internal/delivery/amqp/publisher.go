package amqp

import (
	"context"
	"fmt"
	"time"

	"go-arch/internal/config"
	"go-arch/internal/domain"

	rabbitmq "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn     *Connection
	exchange string
	routes   map[string]string
}

func NewPublisher(conn *Connection, cfg config.RabbitMQConfig) *Publisher {
	return &Publisher{
		conn:     conn,
		exchange: cfg.Exchange,
		routes: map[string]string{
			domain.UserCreatedEvent: cfg.UserCreatedRoutingKey,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, event domain.Event) error {
	routingKey, ok := p.routes[event.Name]
	if !ok {
		return fmt.Errorf("routing key for event %q is not configured", event.Name)
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("open rabbitmq channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Confirm(false); err != nil {
		return fmt.Errorf("enable publish confirm: %w", err)
	}

	confirmations := ch.NotifyPublish(make(chan rabbitmq.Confirmation, 1))
	if err := ch.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		rabbitmq.Publishing{
			ContentType:  "application/json",
			DeliveryMode: rabbitmq.Persistent,
			Timestamp:    time.Now().UTC(),
			Type:         event.Name,
			Body:         event.Payload,
		},
	); err != nil {
		return fmt.Errorf("publish event %q: %w", event.Name, err)
	}

	select {
	case confirmation := <-confirmations:
		if !confirmation.Ack {
			return fmt.Errorf("publish event %q was not acknowledged by broker", event.Name)
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
