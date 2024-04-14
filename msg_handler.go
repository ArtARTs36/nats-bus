package natsbus

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

type natsMessage struct {
	id  string
	msg jetstream.Msg
}

func (b *NatsBus) createMessageHandler(ctx context.Context, consumer *streamConsumer) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		message := &natsMessage{
			id:  msg.Headers().Get(messageHeaderMessageID),
			msg: msg,
		}
		if message.id == "" {
			message.id = generateMessageID()
		}

		slog.
			With(slog.String("message_id", message.id)).
			With(slog.String("topic_name", message.msg.Subject())).
			DebugContext(ctx, "[event-bus] handling message")

		event, err := consumer.subscriber.Event.CreateFromJSON(msg.Data())
		if err != nil {
			slog.
				With(slog.String("err", err.Error())).
				With(slog.String("topic_name", message.msg.Subject())).
				ErrorContext(ctx, "[event-bus] failed to parse payload")

			err = msg.Nak()
			if err != nil {
				slog.
					With(slog.String("err", err.Error())).
					With(slog.String("topic_name", message.msg.Subject())).
					ErrorContext(ctx, "[event-bus] failed to nack message")
				return
			}
			return
		}

		err = consumer.subscriber.Subscriber(event)
		if err != nil {
			slog.
				With(slog.String("err", err.Error())).
				With(slog.String("topic_name", message.msg.Subject())).
				WarnContext(ctx, "[event-bus] failed to handle message")

			nak(ctx, message)
			return
		}

		slog.
			With(slog.String("message_id", message.id)).
			With(slog.String("topic_name", message.msg.Subject())).
			DebugContext(ctx, "[event-bus] message successful handled")

		ack(ctx, message)
	}
}
