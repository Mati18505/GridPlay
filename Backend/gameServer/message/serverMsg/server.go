package serverMsg

import (
	"GridPlay/assert"
	"GridPlay/gameServer/message"
	"fmt"
)

type MsgType message.MsgType
const (
	TMatchStarted MsgType = iota
	TMoveAns
	TOpponentMove
	TWinEvent
	TNotAllowedErr
)

type MatchStarted struct {
	Char rune `json:"char"`
	OpponentChar rune `json:"opponentChar"`
}

type MoveRes struct {
	Approved        bool   `json:"approved"`
	Reason string `json:"reason"`
}

type MoveMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type WinMessage struct {
	Status string `json:"status"`
	Cause string `json:"cause"`
}

type NotAllowedErrMessage struct {
	Reason string `json:"reason"`
}

func (msgT MsgType) String() string { 
	switch msgT {
	case TMatchStarted:
		return "match_started"
	case TMoveAns:
		return "move_answer"
	case TOpponentMove:
		return "opponent_move"
	case TWinEvent:
		return "win_event"
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

func (msg MatchStarted) String() string {
	return fmt.Sprintf("Char: %s OpponentChar: %s", string(msg.Char), string(msg.OpponentChar))
}