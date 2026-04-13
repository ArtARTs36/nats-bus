package natsbus

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

type protoRoundTripEvent struct {
	anypb.Any
}

func (e *protoRoundTripEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "proto.topic",
		SerializationType: SerializationProto,
	}
}

type protoWrongMetaEvent struct {
	anypb.Any
}

func (e *protoWrongMetaEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "proto.topic",
		SerializationType: SerializationJSON,
	}
}

type nonProtoEvent struct{}

func (e nonProtoEvent) TopicMeta() TopicMeta {
	return TopicMeta{
		TopicName:         "proto.topic",
		SerializationType: SerializationProto,
	}
}

func TestProtoSerializer_roundTrip(t *testing.T) {
	t.Parallel()

	ps := NewProtoSerializer()

	ev := &protoRoundTripEvent{
		Any: anypb.Any{
			TypeUrl: "type.googleapis.com/test.Event",
			Value:   []byte("hello"),
		},
	}

	payload, err := ps.Serialize(ev)
	require.NoError(t, err)

	out, err := ps.Deserialize(payload, &protoRoundTripEvent{})
	require.NoError(t, err)

	pe, ok := out.(*protoRoundTripEvent)
	require.True(t, ok)
	require.Equal(t, ev.TypeUrl, pe.TypeUrl)
	require.Equal(t, ev.Value, pe.Value)
}

func TestProtoSerializer_rejectsWrongSerializationType(t *testing.T) {
	t.Parallel()

	ps := NewProtoSerializer()

	_, err := ps.Serialize(&protoWrongMetaEvent{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "SerializationProto")
}

func TestProtoSerializer_serializeRejectsNonProtoMessage(t *testing.T) {
	t.Parallel()

	ps := NewProtoSerializer()

	_, err := ps.Serialize(nonProtoEvent{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "proto.Message")
}

func TestProtoSerializer_deserializeRejectsNonProtoMessagePrototype(t *testing.T) {
	t.Parallel()

	ps := NewProtoSerializer()

	_, err := ps.Deserialize([]byte("abc"), nonProtoEvent{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "prototype must implement proto.Message")
}
