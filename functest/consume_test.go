package functest

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natsbus "github.com/artarts36/nats-bus"
)

type UserEvent struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (e *UserEvent) TopicMeta() natsbus.TopicMeta {
	return natsbus.TopicMeta{
		TopicName:         "users.v1",
		SerializationType: natsbus.SerializationJSON,
	}
}

func TestBus_Consume(t *testing.T) {
	initLogger()

	natsURLs := []string{"nats1:4222", "nats2:4222", "nats3:4222"}
	if os.Getenv("TEST_NATS_URLS") != "" {
		natsURLs = strings.Split(os.Getenv("TEST_NATS_URLS"), ",")
	}

	bus, err := natsbus.Connect(&natsbus.Config{
		ServiceName:             "test",
		NatsURLs:                natsURLs,
		AutoCreateStreamPublish: true,
		CreateStreamTimeout:     10 * time.Second,
		CreateConsumerTimeout:   10 * time.Second,
	}, natsbus.NewJSONSerializer())
	require.NoError(t, err)
	require.NotNil(t, bus)

	defer func() {
		require.NoError(t, bus.Close())
	}()

	var consumedEvent *UserEvent

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bus.Subscribe(natsbus.On(func(_ context.Context, event *UserEvent) error {
		consumedEvent = event

		slog.
			With(slog.Any("event", event)).
			Debug("consumed test event")

		cancel()

		return nil
	}))

	_, err = bus.Publish(ctx, &UserEvent{
		FirstName: "ab",
		LastName:  "cd",
	})

	require.NoError(t, err)

	err = bus.Consume(ctx)

	require.NoError(t, err)

	assert.Equal(t, &UserEvent{
		FirstName: "ab",
		LastName:  "cd",
	}, consumedEvent)
}
