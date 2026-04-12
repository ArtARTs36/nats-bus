package natsbus

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type unknownTypeEvent struct{}

func (unknownTypeEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "x",
		SerializationType: "unknown-format",
	}
}

func TestComposeSerializer_unknownSerializationType(t *testing.T) {
	t.Parallel()

	cs := NewComposeSerializer(map[SerializationType]Serializer{
		SerializationJSON: NewJSONSerializer(),
	})

	_, err := cs.Serialize(unknownTypeEvent{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown serialization type")
}

type jsonRoundTripEvent struct {
	Name string `json:"name"`
}

func (e *jsonRoundTripEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "t",
		SerializationType: SerializationJSON,
	}
}

func TestComposeSerializer_delegatesToJSON(t *testing.T) {
	t.Parallel()

	cs := NewComposeSerializer(map[SerializationType]Serializer{
		SerializationJSON: NewJSONSerializer(),
	})

	ev := &jsonRoundTripEvent{Name: "a"}

	payload, err := cs.Serialize(ev)
	require.NoError(t, err)

	out, err := cs.Deserialize(payload, &jsonRoundTripEvent{})
	require.NoError(t, err)

	je, ok := out.(*jsonRoundTripEvent)
	require.True(t, ok)
	require.Equal(t, "a", je.Name)
}

func TestComposeWithJSON_includesJSONByDefault(t *testing.T) {
	t.Parallel()

	cs := ComposeWithJSON(nil)

	ev := &jsonRoundTripEvent{Name: "b"}

	payload, err := cs.Serialize(ev)
	require.NoError(t, err)

	out, err := cs.Deserialize(payload, &jsonRoundTripEvent{})
	require.NoError(t, err)

	je, ok := out.(*jsonRoundTripEvent)
	require.True(t, ok)
	require.Equal(t, "b", je.Name)
}

type composeCustomEvent struct {
	V string
}

func (e *composeCustomEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "custom.topic",
		SerializationType: "custom-test",
	}
}

func TestComposeWithJSON_mergesCustomSerializer(t *testing.T) {
	t.Parallel()

	const customFormat SerializationType = "custom-test"

	fs, err := NewFuncSerializer(
		func(e Event) ([]byte, error) {
			ce, ok := e.(*composeCustomEvent)
			if !ok {
				return nil, fmt.Errorf("expected *composeCustomEvent, got %T", e)
			}

			return []byte(ce.V), nil
		},
		func(payload []byte, prototype Event) (Event, error) {
			_ = prototype

			return &composeCustomEvent{V: string(payload)}, nil
		},
	)
	require.NoError(t, err)

	cs := ComposeWithJSON(map[SerializationType]Serializer{
		SerializationJSON: NewJSONSerializer(),
		customFormat:      fs,
	})

	ev := &composeCustomEvent{V: "hello"}

	payload, err := cs.Serialize(ev)
	require.NoError(t, err)
	require.Equal(t, "hello", string(payload))

	out, err := cs.Deserialize(payload, &composeCustomEvent{})
	require.NoError(t, err)

	ce, ok := out.(*composeCustomEvent)
	require.True(t, ok)
	require.Equal(t, "hello", ce.V)
}
