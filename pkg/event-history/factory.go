package eventhistory

import (
	"github.com/jmoiron/sqlx"

	natsbus "github.com/artarts36/nats-bus"
)

func Callbacks(db *sqlx.DB) natsbus.Callbacks {
	callbacks := &natsbus.Callbacks{}

	events := NewDBEventRepository(db, DefaultEventsTable)
	updates := NewDBEventUpdateRepository(db, DefaultEventUpdatesTable)

	consumer := NewConsumerHandler(events, updates, false, emptyExtraIDFetcher)
	producer := NewProducerHandler(events, updates, emptyExtraIDFetcher)

	consumer.Register(callbacks)
	producer.Register(callbacks)

	return *callbacks
}

func ConsumerCallbacks(db *sqlx.DB) natsbus.Callbacks {
	callbacks := &natsbus.Callbacks{}

	events := NewDBEventRepository(db, DefaultEventsTable)
	updates := NewDBEventUpdateRepository(db, DefaultEventUpdatesTable)

	consumer := NewConsumerHandler(events, updates, true, emptyExtraIDFetcher)

	consumer.Register(callbacks)

	return *callbacks
}

func ProducerCallbacks(db *sqlx.DB) natsbus.Callbacks {
	callbacks := &natsbus.Callbacks{}

	events := NewDBEventRepository(db, DefaultEventsTable)
	updates := NewDBEventUpdateRepository(db, DefaultEventUpdatesTable)

	producer := NewProducerHandler(events, updates, emptyExtraIDFetcher)

	producer.Register(callbacks)

	return *callbacks
}
