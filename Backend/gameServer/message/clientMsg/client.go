package clientMsg

import (
	"GridPlay/assert"
	"GridPlay/gameServer/message"
)

type MsgType message.MsgType
const (
	TGameMessage MsgType = iota
)

type GameMessage struct {
	Data any `json:"data"`
}

func (msgT MsgType) String() string { 
	switch msgT {
	case TGameMessage:
		return "game_message"
	default:
		assert.Never("unknown type of client message", "client message", msgT)
		return "unknown"
	}
}

func MakeMessage[T any](msgType MsgType, msgData T) message.Message {
	msg := message.MakeMessage(message.MsgType(msgType), msgData)

	return msg
}