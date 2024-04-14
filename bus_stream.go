package natsbus

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"
)

func (b *NatsBus) persistStream(ctx context.Context, stream *natsStream) error {
	createStreamCtx, cancel := context.WithTimeout(ctx, b.cfg.CreateStreamTimeout)
	defer cancel()

	st, err := b.jetStream.CreateOrUpdateStream(createStreamCtx, jetstream.StreamConfig{
		Name:     stream.name,
		Subjects: []string{stream.topic},
	})
	if err != nil {
		return fmt.Errorf("failed to create stream %q: %w", stream.name, err)
	}

	slog.
		With(slog.Any("topic_name", stream.topic)).
		With(slog.String("stream_name", stream.name)).
		DebugContext(ctx, "[event-bus] stream persisted")

	stream.stream = st

	return nil
}

func (b *NatsBus) retrieveStream(topicName string) *natsStream {
	streamName := b.createStreamName(topicName)

	st, exists := b.streams[streamName]
	if !exists {
		st = &natsStream{
			name:  streamName,
			topic: topicName,
		}
		b.streams[streamName] = st
	}

	return st
}
