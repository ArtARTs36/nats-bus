package natsbus

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// JSONSerializer encodes and decodes events as JSON.
type JSONSerializer struct{}

// NewJSONSerializer returns a JSONSerializer.
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

func (j *JSONSerializer) Serialize(event Event) ([]byte, error) {
	if err := j.checkJSONMeta(event.TopicMeta()); err != nil {
		return nil, err
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("natsbus: marshal event to json: %w", err)
	}

	return data, nil
}

func (j *JSONSerializer) Deserialize(payload []byte, prototype Event) (Event, error) {
	if err := j.checkJSONMeta(prototype.TopicMeta()); err != nil {
		return nil, err
	}

	dst, err := jsonDecodeTarget(prototype)
	if err != nil {
		return nil, err
	}

	if unmarshalErr := json.Unmarshal(payload, dst.Interface()); unmarshalErr != nil {
		return nil, fmt.Errorf("natsbus: unmarshal json into event: %w", unmarshalErr)
	}

	ev, ok := dst.Interface().(Event)
	if !ok {
		return nil, fmt.Errorf("natsbus: decoded value does not implement Event")
	}

	return ev, nil
}

func (j *JSONSerializer) checkJSONMeta(meta TopicMeta) error {
	if meta.SerializationType != SerializationJSON {
		return fmt.Errorf("natsbus: JSONSerializer expects SerializationJSON, got %q", meta.SerializationType)
	}

	return nil
}

func jsonDecodeTarget(prototype Event) (reflect.Value, error) {
	pv := reflect.ValueOf(prototype)
	if !pv.IsValid() {
		return reflect.Value{}, fmt.Errorf("natsbus: nil event prototype")
	}

	t := pv.Type()

	switch t.Kind() {
	case reflect.Ptr:
		if t.Elem().Kind() != reflect.Struct {
			return reflect.Value{}, fmt.Errorf("natsbus: JSON event prototype must be pointer to struct")
		}

		return reflect.New(t.Elem()), nil
	case reflect.Struct:
		return reflect.New(t), nil
	default:
		return reflect.Value{}, fmt.Errorf("natsbus: JSON event prototype must be struct or pointer to struct")
	}
}
