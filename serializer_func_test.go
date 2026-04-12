package natsbus

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFuncSerializer_nilCallbacks(t *testing.T) {
	t.Parallel()

	_, err := NewFuncSerializer(nil, func([]byte, Event) (Event, error) {
		return nil, errors.New("deserialize stub")
	})
	require.Error(t, err)

	_, err = NewFuncSerializer(func(Event) ([]byte, error) { return nil, nil }, nil)
	require.Error(t, err)
}

func TestFuncSerializer_wrapsJSONSerializer(t *testing.T) {
	t.Parallel()

	js := NewJSONSerializer()

	fs, err := NewFuncSerializer(js.Serialize, js.Deserialize)
	require.NoError(t, err)

	ev := &jsonRoundTripEvent{Name: "wrapped"}

	payload, err := fs.Serialize(ev)
	require.NoError(t, err)

	out, err := fs.Deserialize(payload, &jsonRoundTripEvent{})
	require.NoError(t, err)

	je, ok := out.(*jsonRoundTripEvent)
	require.True(t, ok)
	require.Equal(t, "wrapped", je.Name)
}
