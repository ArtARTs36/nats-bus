package natsbus

import (
	"context"
	"log/slog"
)

func nak(ctx context.Context, msg *natsMessage) {
	err := msg.msg.Nak()
	if err == nil {
		return
	}

	slog.
		With(slog.Any("err", err)).
		With(slog.String("message_id", msg.id)).
		With(slog.String("topic_name", msg.msg.Subject())).
		ErrorContext(ctx, "[nats-bus] failed to nak")
}

func ack(ctx context.Context, msg *natsMessage) {
	err := msg.msg.Ack()
	if err == nil {
		slog.
			With(slog.String("message_id", msg.id)).
			With(slog.String("topic_name", msg.msg.Subject())).
			DebugContext(ctx, "[nats-bus] message ack sent")

		return
	}

	slog.
		With(slog.Any("err", err)).
		With(slog.String("message_id", msg.id)).
		With(slog.String("topic_name", msg.msg.Subject())).
		ErrorContext(ctx, "[nats-bus] failed to ack")
}
