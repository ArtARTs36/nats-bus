package functest

import (
	"context"
	"encoding/json"
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

func (e UserEvent) TopicName() string {
	return "users.v1"
}

func (e UserEvent) CreateFromJSON(data []byte) (natsbus.Event, error) {
	var event UserEvent

	err := json.Unmarshal(data, &event)

	return &event, err
}

func TestBus_Consume(t *testing.T) {
	initLogger()

	natsURLs := []string{"nats1:4222", "nats2:4222", "nats3:4222"}
	if os.Getenv("TEST_NATS_URLS") != "" {
		natsURLs = strings.Split(os.Getenv("TEST_NATS_URLS"), "")
	}

	bus, err := natsbus.Connect(&natsbus.Config{
		ServiceName:             "test",
		NatsURLs:                natsURLs,
		AutoCreateStreamPublish: true,
		CreateStreamTimeout:     10 * time.Second,
		CreateConsumerTimeout:   10 * time.Second,
	})
	defer func() {
		err := bus.Close()
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	require.NotNil(t, bus)

	var consumedEvent *natsbus.ConsumedEvent

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bus.Subscribe(&natsbus.EventSubscriber{
		Event: UserEvent{},
		Subscriber: func(event *natsbus.ConsumedEvent) error {
			consumedEvent = event

			slog.
				With(slog.String("timestamp", event.Timestamp.Format(time.DateTime))).
				With(slog.String("id", event.ID)).
				Debug("consumed test event")

			cancel()

			return nil
		},
	})

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
	}, consumedEvent.Event)
}
