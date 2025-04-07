package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalMessage(t *testing.T) {
	raw := `{"type":0,"data":{"x":1,"y":2}}`
	msg, err := UnmarshalMessage([]byte(raw))

	assert.NoError(t, err, "UnmarshalMessage should not return an error")
	assert.Equal(t, TMove, msg.Type, "Message type should be TMove")

	data, ok := msg.Data.(map[string]any)
	assert.True(t, ok, "Message data should be a map")
	assert.Equal(t, float64(1), data["x"], "X should be 1")
	assert.Equal(t, float64(2), data["y"], "Y should be 2")
}

func TestCreateHeader(t *testing.T) {
	header := CreateHeader(TMove)
	assert.Equal(t, TMove, header.Type, "Header type should be TMove")
}

func TestMsgTypeString(t *testing.T) {
	assert.Equal(t, "move", TMove.String(), "String representation of TMove should be 'move'")
}

func TestExtractConcreteMessage(t *testing.T) {
	raw := `{"type":0,"data":{"x":1,"y":2}}`
	msg, _ := UnmarshalMessage([]byte(raw))

	concrete, err := GetConcreteMessage[MoveMessage](msg)
	assert.NoError(t, err, "ExtractConcreteMessage should not return an error")
	assert.Equal(t, 1, concrete.X, "X should be 1")
	assert.Equal(t, 2, concrete.Y, "Y should be 2")
}

func TestExtractConcreteMessageWrongType(t *testing.T) {
	raw := `{"type":0,"data":{"x":1,"y":2}}`
	msg, _ := UnmarshalMessage([]byte(raw))

	_, err := GetConcreteMessage[string](msg)
	assert.Error(t, err, "ExtractConcreteMessage should return an error for wrong type")
}
