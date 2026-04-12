package natsbus

// Serializer encodes events for JetStream and decodes payloads back into concrete event values.
type Serializer interface {
	Serialize(event Event) ([]byte, error)
	Deserialize(payload []byte, prototype Event) (Event, error)
}
