package serverMsg

import (
	"GridPlay/assert"
	"GridPlay/gameServer/message"
)

type MsgType message.MsgType
const (
	TGameEnded MsgType = iota
	TGameMessage
	TApprove
	TNotAllowedErr
)

type GameEnded struct {
	Status string `json:"status"`
	Cause string `json:"cause"`
}

type GameMessage struct {
	Name any `json:"data"`
	Data any `json:"data"`
}

type Approve struct {
	Approved bool   `json:"approved"`
	Reason string `json:"reason"`
}

type NotAllowedErr struct {
	Reason string `json:"reason"`
}

func (msgT MsgType) String() string { 
	switch msgT {
	case TGameEnded:
		return "game_ended"
	case TGameMessage:
		return "game_message"
	case TApprove:
		return "approve"
	case TNotAllowedErr:
		return "not_allowed_error"
	default:
		assert.Never("unknown type of server message", "server message", msgT)
		return "unknown"
	}
}

func MakeMessage[T any](msgType MsgType, msgData T) message.Message {
	msg := message.MakeMessage(message.MsgType(msgType), msgData)

	return msg
}