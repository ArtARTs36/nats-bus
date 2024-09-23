package eventhistory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	natsbus "github.com/artarts36/nats-bus"
)

type ConsumerHandler struct {
	events                EventRepository
	updates               EventUpdateRepository
	createEventOnHandling bool
	extraIDFetcher        ExtraIDFetcher
}

func NewConsumerHandler(
	events EventRepository,
	updates EventUpdateRepository,
	createEventOnHandling bool,
	extraIDFetcher ExtraIDFetcher,
) *ConsumerHandler {
	return &ConsumerHandler{
		events:                events,
		updates:               updates,
		createEventOnHandling: createEventOnHandling,
		extraIDFetcher:        extraIDFetcher,
	}
}

func (h *ConsumerHandler) Register(callbacks *natsbus.Callbacks) {
	callbacks.Consumer.OnHandling = h.OnHandling
	callbacks.Consumer.OnSucceed = h.OnSucceed
	callbacks.Consumer.OnFailed = h.OnFailed
}

func (h *ConsumerHandler) OnHandling(ctx context.Context, event *natsbus.ConsumedEvent) {
	if h.createEventOnHandling {
		if err := h.createEvent(ctx, event); err != nil {
			slog.
				With(slog.Any("err", err)).
				ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event")

			return
		}
	}

	if err := h.createEventUpdate(ctx, event, EventStatusConsuming, ""); err != nil {
		slog.
			With(slog.Any("err", err)).
			ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event update")

		return
	}
}

func (h *ConsumerHandler) OnSucceed(ctx context.Context, event *natsbus.ConsumedEvent) {
	if err := h.createEventUpdate(ctx, event, EventStatusConsumed, ""); err != nil {
		slog.
			With(slog.Any("err", err)).
			ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event update")

		return
	}
}

func (h *ConsumerHandler) OnFailed(ctx context.Context, event *natsbus.ConsumedEvent, consumeErr error) {
	if err := h.createEventUpdate(ctx, event, EventStatusConsumeFailed, consumeErr.Error()); err != nil {
		slog.
			With(slog.Any("err", err)).
			ErrorContext(ctx, "[nats-bus][callback-handler] failed to create event update")

		return
	}
}

func (h *ConsumerHandler) createEvent(ctx context.Context, event *natsbus.ConsumedEvent) error {
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

func (h *ConsumerHandler) createEventUpdate(
	ctx context.Context,
	event *natsbus.ConsumedEvent,
	status EventStatus,
	description string,
) error {
	model := &EventUpdate{
		EventID: event.ID,
		Status:  status,
		Description: sql.NullString{
			Valid:  description != "",
			String: description,
		},
		CreatedAt: time.Now(),
	}

	return h.updates.Create(ctx, model)
}
