package serverMsg

import (
	"GridPlay/assert"
	"GridPlay/gameServer/message"
	"fmt"
)

type MsgType message.MsgType
const (
	TGameStarted MsgType = iota
	TGameEnded
	TGameMessage
	TApprove
	TNotAllowedErr
)

type GameStarted struct {
	Char rune `json:"char"`
	OpponentChar rune `json:"opponentChar"`
}

type GameEnded struct {
	Status string `json:"status"`
	Cause string `json:"cause"`
}

type GameMessage struct {
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
	case TGameStarted:
		return "game_started"
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

func (msg GameStarted) String() string {
	return fmt.Sprintf("Char: %s OpponentChar: %s", string(msg.Char), string(msg.OpponentChar))
}