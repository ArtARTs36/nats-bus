package eventhistory

import (
	"database/sql"
	"time"
)

type EventStatus string

const (
	EventStatusProduced      EventStatus = "produced"
	EventStatusConsuming     EventStatus = "consuming"
	EventStatusConsumed      EventStatus = "consumed"
	EventStatusConsumeFailed EventStatus = "consume_failed"

	DefaultEventsTable       = "events"
	DefaultEventUpdatesTable = "event_updates"
)

type Event struct {
	ID        string         `db:"id"`
	ExtraID   sql.NullString `db:"extra_id"`
	Topic     string         `db:"topic"`
	Payload   []byte         `db:"payload"`
	CreatedAt time.Time      `db:"created_at"`
}

type EventUpdate struct {
	EventID     string         `db:"event_id"`
	Status      EventStatus    `db:"status"`
	Description sql.NullString `db:"description"`
	CreatedAt   time.Time      `db:"created_at"`
}
