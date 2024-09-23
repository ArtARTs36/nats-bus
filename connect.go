package natsbus

import (
	"context"
	"errors"
	"fmt"

	oNats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func Connect(natsCfg *Config, callbacks Callbacks) (*NatsBus, error) {
	conn, err := connect(natsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	prepareBusCallbacks(&callbacks)

	return &NatsBus{
		cfg:       natsCfg,
		conn:      conn,
		jetStream: js,
		streams:   map[string]*natsStream{},
		consumers: []*streamConsumer{},
		callbacks: callbacks,
	}, nil
}

func prepareBusCallbacks(callbacks *Callbacks) {
	if callbacks.Consumer.OnHandling != nil {
		callbacks.Consumer.OnHandling = noopConsumerCallback
	}
	if callbacks.Consumer.OnSucceed != nil {
		callbacks.Consumer.OnSucceed = noopConsumerCallback
	}
	if callbacks.Consumer.OnFailed != nil {
		callbacks.Consumer.OnFailed = func(_ context.Context, _ *ConsumedEvent, _ error) {}
	}
	if callbacks.Producer.OnProduced != nil {
		callbacks.Producer.OnProduced = func(ctx context.Context, event *ProducedEvent) {}
	}
}

func connect(natsCfg *Config) (*oNats.Conn, error) {
	errs := make([]error, 0)

	for _, url := range natsCfg.NatsURLs {
		conn, err := oNats.Connect(url)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		return conn, nil
	}

	return nil, errors.Join(errs...)
}
