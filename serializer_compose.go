package natsbus

import "fmt"

// ComposeSerializer delegates to a Serializer chosen by TopicMeta.SerializationType.
type ComposeSerializer struct {
	byType map[SerializationType]Serializer
}

// NewComposeSerializer builds a serializer that routes by serialization type.
func NewComposeSerializer(byType map[SerializationType]Serializer) *ComposeSerializer {
	return &ComposeSerializer{byType: byType}
}

func (c *ComposeSerializer) Serialize(event Event) ([]byte, error) {
	meta := event.TopicMeta()
	s, ok := c.byType[meta.SerializationType]
	if !ok {
		return nil, fmt.Errorf("natsbus: unknown serialization type %q", meta.SerializationType)
	}

	return s.Serialize(event)
}

func (c *ComposeSerializer) Deserialize(payload []byte, prototype Event) (Event, error) {
	meta := prototype.TopicMeta()
	s, ok := c.byType[meta.SerializationType]
	if !ok {
		return nil, fmt.Errorf("natsbus: unknown serialization type %q", meta.SerializationType)
	}

	return s.Deserialize(payload, prototype)
}
