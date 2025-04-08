package clientMsg

import (
	"GridPlay/assert"
	"GridPlay/gameServer/message"
)

type MsgType message.MsgType
const (
	TMove MsgType = iota
)

type MoveMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
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

func MakeMessage[T any](msgType MsgType, msgData T) message.Message {
	msg := message.MakeMessage(message.MsgType(msgType), msgData)

	return msg
}