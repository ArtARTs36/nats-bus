package natsbus

import (
	"context"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type Bus interface {
	Publish(ctx context.Context, event Event) (*ProducedEvent, error)
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

type ProducedEvent struct {
	ID string
}

func (s *EventSubscriber) toConsumerConfig(name string) jetstream.ConsumerConfig {
	cfg := jetstream.ConsumerConfig{
		Durable:       name,
		DeliverPolicy: jetstream.DeliverLastPolicy,
	}

	if s.MaxAttempts != nil {
		cfg.MaxDeliver = *s.MaxAttempts
	}

	return cfg
}
