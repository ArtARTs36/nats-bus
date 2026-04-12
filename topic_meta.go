package natsbus

// SerializationType identifies how an event payload is encoded on the wire.
type SerializationType string

const (
	// SerializationJSON encodes events as JSON.
	SerializationJSON SerializationType = "json"
)

// TopicMeta describes where and how an event is published.
type TopicMeta struct {
	TopicName         string
	SerializationType SerializationType
}
