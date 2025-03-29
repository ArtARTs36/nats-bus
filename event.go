package natsbus

import (
	"context"
	"reflect"
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
	Subscriber  func(ctx context.Context, event *ConsumedEvent) error
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

func On[E Event](
	subscriber func(ctx context.Context, event E) error,
) *EventSubscriber {
	return &EventSubscriber{
		Event: func() Event {
			var eventInstance E

			value := reflect.ValueOf(eventInstance)
			if value.Type().Kind() == reflect.Ptr {
				v := reflect.New(reflect.TypeOf(eventInstance).Elem())

				return v.Interface().(Event)
			}

			return eventInstance
		}(),
		Subscriber: func(ctx context.Context, event *ConsumedEvent) error {
			return subscriber(ctx, event.Event.(E))
		},
	}
}
