package amqp

import (
	"context"

	"go-arch/internal/config"

	rabbitmq "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
)

type ConsumerService struct {
	conn   *Connection
	cfg    config.RabbitMQConfig
	logger *zerolog.Logger
}

func NewConsumerService(conn *Connection, cfg config.RabbitMQConfig, logger *zerolog.Logger) *ConsumerService {
	return &ConsumerService{
		conn:   conn,
		cfg:    cfg,
		logger: logger,
	}
}

func (s *ConsumerService) Start(ctx context.Context) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}

	if err := ch.Qos(10, 0, false); err != nil {
		ch.Close()
		return err
	}

	deliveries, err := ch.Consume(
		s.cfg.UserCreatedQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return err
	}

	go func() {
		defer ch.Close()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info().Msg("rabbitmq consumer stopped")
				return
			case delivery, ok := <-deliveries:
				if !ok {
					s.logger.Warn().Msg("rabbitmq deliveries channel closed")
					return
				}

				s.handleUserCreated(ctx, delivery)
			}
		}
	}()

	return nil
}

func (s *ConsumerService) handleUserCreated(ctx context.Context, delivery rabbitmq.Delivery) {
	s.logger.Info().
		Str("routing_key", delivery.RoutingKey).
		Str("type", delivery.Type).
		RawJSON("payload", delivery.Body).
		Msg("user created event received")

	if err := delivery.Ack(false); err != nil {
		s.logger.Error().Err(err).Msg("ack rabbitmq delivery")
	}
}
