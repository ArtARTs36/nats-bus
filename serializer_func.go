package natsbus

import "fmt"

// FuncSerializer implements Serializer by delegating to user-supplied functions.
// Use it with [Connect] or inside [NewComposeSerializer] / [ComposeWithJSON] when you do not want a named type.
type FuncSerializer struct {
	SerializeFunc   func(event Event) ([]byte, error)
	DeserializeFunc func(payload []byte, prototype Event) (Event, error)
}

// NewFuncSerializer wraps serialize and deserialize callbacks as a [Serializer].
// Both functions must be non-nil.
func NewFuncSerializer(
	serialize func(event Event) ([]byte, error),
	deserialize func(payload []byte, prototype Event) (Event, error),
) (*FuncSerializer, error) {
	if serialize == nil {
		return nil, fmt.Errorf("natsbus: FuncSerializer: serialize function is required")
	}

	if deserialize == nil {
		return nil, fmt.Errorf("natsbus: FuncSerializer: deserialize function is required")
	}

	return &FuncSerializer{
		SerializeFunc:   serialize,
		DeserializeFunc: deserialize,
	}, nil
}

func (f *FuncSerializer) Serialize(event Event) ([]byte, error) {
	return f.SerializeFunc(event)
}

func (f *FuncSerializer) Deserialize(payload []byte, prototype Event) (Event, error) {
	return f.DeserializeFunc(payload, prototype)
}
