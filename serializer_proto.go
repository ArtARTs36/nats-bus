package natsbus

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"
)

// ProtoSerializer encodes and decodes events as protobuf binary payload.
type ProtoSerializer struct{}

// NewProtoSerializer returns a ProtoSerializer.
func NewProtoSerializer() *ProtoSerializer {
	return &ProtoSerializer{}
}

func (p *ProtoSerializer) Serialize(event Event) ([]byte, error) {
	if err := p.checkProtoMeta(event.TopicMeta()); err != nil {
		return nil, err
	}

	message, ok := event.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("natsbus: proto event must implement proto.Message, got %T", event)
	}

	data, err := proto.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("natsbus: marshal event to proto: %w", err)
	}

	return data, nil
}

func (p *ProtoSerializer) Deserialize(payload []byte, prototype Event) (Event, error) {
	if err := p.checkProtoMeta(prototype.TopicMeta()); err != nil {
		return nil, err
	}

	dst, err := protoDecodeTarget(prototype)
	if err != nil {
		return nil, err
	}

	message, ok := dst.Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("natsbus: proto event prototype must implement proto.Message")
	}

	if unmarshalErr := proto.Unmarshal(payload, message); unmarshalErr != nil {
		return nil, fmt.Errorf("natsbus: unmarshal proto into event: %w", unmarshalErr)
	}

	event, ok := dst.Interface().(Event)
	if !ok {
		return nil, fmt.Errorf("natsbus: decoded value does not implement Event")
	}

	return event, nil
}

func (p *ProtoSerializer) checkProtoMeta(meta TopicMeta) error {
	if meta.SerializationType != SerializationProto {
		return fmt.Errorf("natsbus: ProtoSerializer expects SerializationProto, got %q", meta.SerializationType)
	}

	return nil
}

func protoDecodeTarget(prototype Event) (reflect.Value, error) {
	pv := reflect.ValueOf(prototype)
	if !pv.IsValid() {
		return reflect.Value{}, fmt.Errorf("natsbus: nil event prototype")
	}

	t := pv.Type()

	switch t.Kind() { //nolint:exhaustive // not need
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("natsbus: proto event prototype must be pointer to struct")
		}

		return reflect.New(t.Elem()), nil
	case reflect.Struct:
		return reflect.New(t), nil
	default:
		return reflect.Value{}, fmt.Errorf("natsbus: proto event prototype must be struct or pointer to struct")
	}
}
