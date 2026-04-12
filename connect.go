package natsbus

import (
	"errors"
	"fmt"

	oNats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func Connect(natsCfg *Config, serializer Serializer) (*NatsBus, error) {
	if serializer == nil {
		return nil, fmt.Errorf("serializer is required")
	}

	conn, err := connect(natsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	return &NatsBus{
		cfg:        natsCfg,
		serializer: serializer,
		conn:       conn,
		jetStream:  js,
		streams:    map[string]*natsStream{},
		consumers:  []*streamConsumer{},
	}, nil
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
