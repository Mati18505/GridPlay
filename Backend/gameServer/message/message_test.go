package message

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapMessage(t *testing.T) {
	header := CreateHeader(1)
	data := map[string]any{"key": "value"}
	msg := WrapMessage(header, data)

	assert.Equal(t, header, msg.MessageHeader)
	assert.Equal(t, data, msg.Data)
}

func TestMarshalMessage(t *testing.T) {
	msg := Message{
		MessageHeader: MessageHeader{Type: 1},
		Data:          map[string]any{"key": "value"},
	}
	bytes := msg.MarshalMessage()

	var unmarshalled Message
	err := json.Unmarshal(bytes, &unmarshalled)

	assert.NoError(t, err)
	assert.Equal(t, msg, unmarshalled)
}

func TestUnmarshalMessage(t *testing.T) {
	raw := `{"type":1,"data":{"key":"value"}}`
	msg, err := UnmarshalMessage([]byte(raw))

	assert.NoError(t, err)
	assert.Equal(t, MsgType(1), msg.Type)
	assert.Equal(t, map[string]any{"key": "value"}, msg.Data)
}

func TestMakeMessage(t *testing.T) {
	data := map[string]any{"key": "value"}
	msg := MakeMessage(1, data)

	assert.Equal(t, MsgType(1), msg.Type)
	assert.Equal(t, data, msg.Data)
}

func TestGetConcreteMessage(t *testing.T) {
	raw := `{"type":1,"data":{"key":"value"}}`
	msg, _ := UnmarshalMessage([]byte(raw))

	concrete, err := GetConcreteMessage[map[string]any](msg)
	assert.NoError(t, err)
	assert.Equal(t, "value", concrete["key"])
}

func TestGetConcreteMessageWrongType(t *testing.T) {
	raw := `{"type":1,"data":{"key":"value"}}`
	msg, _ := UnmarshalMessage([]byte(raw))

	_, err := GetConcreteMessage[string](msg)
	assert.Error(t, err)
}
