package natsbus

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type natsMessage struct {
	id        string
	msg       jetstream.Msg
	timestamp time.Time
}

func (b *NatsBus) createMessageHandler(ctx context.Context, consumer *streamConsumer) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		message := &natsMessage{
			id:  msg.Headers().Get(messageHeaderMessageID),
			msg: msg,
		}

		md, err := msg.Metadata()
		if err != nil {
			slog.
				With(slog.String("err", err.Error())).
				WarnContext(ctx, "[nats-bus] failed to get metadata from message")
		} else {
			message.timestamp = md.Timestamp
		}

		if message.id == "" {
			message.id = generateMessageID()
		}

		slog.
			With(slog.String("message_id", message.id)).
			With(slog.String("topic_name", message.msg.Subject())).
			DebugContext(ctx, "[nats-bus] handling message")

		event, err := consumer.subscriber.Event.CreateFromJSON(msg.Data())
		if err != nil {
			slog.
				With(slog.String("err", err.Error())).
				With(slog.String("topic_name", message.msg.Subject())).
				ErrorContext(ctx, "[nats-bus] failed to parse payload")

			err = msg.Nak()
			if err != nil {
				slog.
					With(slog.String("err", err.Error())).
					With(slog.String("topic_name", message.msg.Subject())).
					ErrorContext(ctx, "[nats-bus] failed to nack message")
				return
			}
			return
		}

		err = consumer.subscriber.Subscriber(&ConsumedEvent{
			Event:     event,
			ID:        message.id,
			Timestamp: message.timestamp,
		})
		if err != nil {
			slog.
				With(slog.String("err", err.Error())).
				With(slog.String("topic_name", message.msg.Subject())).
				WarnContext(ctx, "[nats-bus] failed to handle message")

			nak(ctx, message)
			return
		}

		slog.
			With(slog.String("message_id", message.id)).
			With(slog.String("topic_name", message.msg.Subject())).
			DebugContext(ctx, "[nats-bus] message successful handled")

		ack(ctx, message)
	}
}
