package server

import (
	"TicTacToe/assert"
	"encoding/json"
)

type MsgType int
const (
	TMatchStarted MsgType = iota
	TMoveAns
	TOpponentMove
	TWinEvent
	TNotAllowedErr
)

type MessageHeader struct {
	Type MsgType `json:"type"`
}

type Message struct {
	MessageHeader
	Data any `json:"data"`
}

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

func (msg Message) MarshalMessage() []byte {
	bytes, err := json.Marshal(&msg)
	assert.NoError(err, "cannot marshal message")

	return bytes
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

func MakeMessage[T any](msgType MsgType, msgData T) Message {
	header := CreateHeader(msgType)
	msg := WrapMessage(header, msgData)

	return msg
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