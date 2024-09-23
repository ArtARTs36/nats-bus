package eventhistory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	natsbus "github.com/artarts36/nats-bus"
)

type ProducerHandler struct {
	events         EventRepository
	updates        EventUpdateRepository
	extraIDFetcher ExtraIDFetcher
}

func NewProducerHandler(
	events EventRepository,
	updates EventUpdateRepository,
	extraIDFetcher ExtraIDFetcher,
) *ProducerHandler {
	return &ProducerHandler{
		events:         events,
		updates:        updates,
		extraIDFetcher: extraIDFetcher,
	}
}

func (h *ProducerHandler) Register(callbacks *natsbus.Callbacks) {
	callbacks.Producer.OnProduced = h.OnProduced
}

func (h *ProducerHandler) OnProduced(ctx context.Context, event *natsbus.ProducedEvent) {
	if err := h.createEvent(ctx, event); err != nil {
		slog.
			With(slog.Any("err", err)).
			ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event")

		return
	}

	if err := h.createEventUpdate(ctx, event, ""); err != nil {
		slog.
			With(slog.Any("err", err)).
			ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event update")

		return
	}
}

func (h *ProducerHandler) createEvent(ctx context.Context, event *natsbus.ProducedEvent) error {
	payload, err := json.Marshal(event.Event)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	extraID, extraIDOk := h.extraIDFetcher(event.Event)

	model := &Event{
		ID: event.ID,
		ExtraID: sql.NullString{
			Valid:  extraIDOk,
			String: extraID,
		},
		Topic:     event.Event.TopicName(),
		Payload:   payload,
		CreatedAt: event.Timestamp,
	}

	return h.events.Create(ctx, model)
}

func (h *ProducerHandler) createEventUpdate(
	ctx context.Context,
	event *natsbus.ProducedEvent,
	description string,
) error {
	model := &EventUpdate{
		EventID: event.ID,
		Status:  EventStatusProduced,
		Description: sql.NullString{
			Valid:  description != "",
			String: description,
		},
		CreatedAt: event.Timestamp,
	}

	return h.updates.Create(ctx, model)
}
