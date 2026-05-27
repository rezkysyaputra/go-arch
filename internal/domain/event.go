package domain

import "context"

const UserCreatedEvent = "user.created"

type Event struct {
	Name    string
	Payload []byte
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}
