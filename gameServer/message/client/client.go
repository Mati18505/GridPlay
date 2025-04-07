package client

import (
	"TicTacToe/assert"
	"encoding/json"
	"errors"
)

type MsgType int
const (
	TMove MsgType = iota
)

type MessageHeader struct {
	Type MsgType `json:"type"`
}

type Message struct {
	MessageHeader
	Data any `json:"data"`
}

type MoveMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func UnmarshalMessage(bytes []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(bytes, &msg)

	if err != nil {
		return msg, errors.New("can't unmarshal message")
	}
	
	return msg, nil
}

func CreateHeader(msgType MsgType) MessageHeader {
	msgHeader := MessageHeader{
		Type: msgType,
	}

	return msgHeader
}

func (msgT MsgType) String() string { 
	switch msgT {
	case TMove:
		return "move"
	default:
		assert.Never("unknown type of client message", "client message", msgT)
		return "unknown"
	}
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