package natsbus

import (
	"context"
	"time"
)

type Bus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(subscriber *EventSubscriber)
	Consume(ctx context.Context) error
	Close() error
}

type Event interface {
	TopicName() string
	CreateFromJSON([]byte) (Event, error)
}

type EventSubscriber struct {
	Event       Event
	Subscriber  func(event *ConsumedEvent) error
	MaxAttempts *int
}

type ConsumedEvent struct {
	Event     Event
	ID        string
	Timestamp time.Time
}
