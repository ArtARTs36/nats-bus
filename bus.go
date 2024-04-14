package natsbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	oNats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsBus struct {
	cfg       *Config
	conn      *oNats.Conn
	jetStream jetstream.JetStream

	streams   map[string]*natsStream
	consumers []*streamConsumer
}

type streamConsumer struct {
	name string

	stream     *natsStream
	subscriber *EventSubscriber

	consumer        jetstream.Consumer
	consumerContext jetstream.ConsumeContext
}

type natsStream struct {
	name  string
	topic string

	stream jetstream.Stream
}

type Config struct {
	ServiceName             string        `env:"SERVICE_NAME,required"`
	NatsURLs                []string      `env:"URLS,required"`
	CreateStreamTimeout     time.Duration `env:"CREATE_STREAM_TIMEOUT" envDefault:"10s"`
	CreateConsumerTimeout   time.Duration `env:"CREATE_CONSUMER_TIMEOUT" envDefault:"10s"`
	AutoCreateStreamPublish bool          `env:"AUTO_CREATE_STREAM_ON_PUBLISH"`
}

func (b *NatsBus) Publish(ctx context.Context, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal Event to json: %w", err)
	}

	msg := oNats.NewMsg(event.TopicName())
	msg.Header.Set(messageHeaderMessageID, generateMessageID())
	msg.Data = data

	_, err = b.jetStream.PublishMsg(ctx, msg)
	if err == nil {
		return nil
	}

	if !b.cfg.AutoCreateStreamPublish && !errors.Is(err, jetstream.ErrNoStreamResponse) {
		return err
	}

	st := b.retrieveStream(event.TopicName())
	if st.stream == nil {
		twoErr := b.persistStream(ctx, st)
		if twoErr != nil {
			return errors.Join(err, twoErr)
		}

		_, twoErr = b.jetStream.PublishMsg(ctx, msg)
		if twoErr != nil {
			return errors.Join(err, twoErr)
		}
		err = nil
	}

	return err
}

func (b *NatsBus) Subscribe(subscriber *EventSubscriber) {
	fillSubscriber(subscriber)

	st := b.retrieveStream(subscriber.Event.TopicName())

	b.consumers = append(b.consumers, &streamConsumer{
		name:       b.createConsumerName(st.name),
		stream:     st,
		subscriber: subscriber,
	})
}

func (b *NatsBus) Close() error {
	slog.Info("[event-bus] closing")

	for _, c := range b.consumers {
		if c.consumerContext != nil {
			c.consumerContext.Stop()
		}
	}

	return b.conn.Drain()
}

func (b *NatsBus) Consume(ctx context.Context) error {
	slog.DebugContext(ctx, "[event-bus] start consuming")

	for _, cons := range b.consumers {
		if cons.stream == nil {
			err := b.persistStream(ctx, cons.stream)
			if err != nil {
				return err
			}
		}

		createConsumerCtx, cancel := context.WithTimeout(ctx, b.cfg.CreateConsumerTimeout)

		consumer, err := b.jetStream.CreateOrUpdateConsumer(
			createConsumerCtx,
			cons.stream.name,
			jetstream.ConsumerConfig{
				Durable:       cons.name,
				DeliverPolicy: jetstream.DeliverLastPolicy,
				MaxDeliver:    *cons.subscriber.MaxAttempts,
			},
		)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to create consumer for consumer %q: %w", cons.name, err)
		}

		cons.consumer = consumer
	}

	wg := &sync.WaitGroup{}

	for _, consumer := range b.consumers {
		c := consumer

		wg.Add(1)

		go func() {
			consCtx, err := c.consumer.Consume(b.createMessageHandler(ctx, c))

			if err != nil {
				wg.Done()

				slog.
					With(slog.String("err", err.Error())).
					ErrorContext(ctx, "[event-bus] failed to consume message")

				return
			}

			slog.
				With(slog.String("topic_name", c.stream.topic)).
				InfoContext(ctx, "[event-bus] consumer started")

			c.consumerContext = consCtx

			for { //nolint:gosimple // not need
				select {
				case <-ctx.Done():
					wg.Done()

					slog.
						With(slog.String("topic_name", c.stream.topic)).
						InfoContext(ctx, "[event-bus] stopping consumer")

					c.consumerContext.Stop()

					return
				}
			}
		}()
	}

	wg.Wait()

	return nil
}

func (b *NatsBus) createConsumerName(streamName string) string {
	return fmt.Sprintf(
		"%s_%s_consumer",
		b.cfg.ServiceName,
		streamName,
	)
}

func (b *NatsBus) createStreamName(topic string) string {
	return fmt.Sprintf(
		"topic_%s",
		strings.ReplaceAll(topic, ".", "_"),
	)
}

func fillSubscriber(subscriber *EventSubscriber) {
	const defaultMaxAttempts = 3

	if subscriber.MaxAttempts == nil {
		a := defaultMaxAttempts
		subscriber.MaxAttempts = &a
	}
}