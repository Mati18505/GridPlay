package message

import (
	"GridPlay/assert"
	"encoding/json"
	"errors"
)

type MsgType int
type MessageHeader struct {
	Type MsgType `json:"type"`
}

type Message struct {
	MessageHeader
	Data any `json:"data"`
}

func CreateHeader(msgType MsgType) MessageHeader {
	msgHeader := MessageHeader{
		Type: msgType,
	}

	return msgHeader
}

func WrapMessage(header MessageHeader, data any) Message {
	return Message{
		MessageHeader: header,
		Data: data,
	}
}

func (msg Message) MarshalMessage() []byte {
	bytes, err := json.Marshal(&msg)
	assert.NoError(err, "cannot marshal message")

	return bytes
}

func UnmarshalMessage(bytes []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(bytes, &msg)

	if err != nil {
		return msg, errors.New("can't unmarshal message")
	}
	
	return msg, nil
}

func MakeMessage[T any](msgType MsgType, msgData T) Message {
	header := CreateHeader(msgType)
	msg := WrapMessage(header, msgData)

	return msg
}

func GetConcreteMessage[T any](message Message) (T, error) {
	var concrete T
	dataBytes, err := json.Marshal(message.Data)
	if err != nil {
		return concrete, errors.New("failed to marshal message data")
	}

	err = json.Unmarshal(dataBytes, &concrete)
	if err != nil {
		return concrete, errors.New("failed to unmarshal message data into concrete type")
	}

	return concrete, nil
}