package message

import (
	"TicTacToe/assert"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type Message struct {
	Type int `json:"type"`
	Data json.RawMessage `json:"data"`
}

type MoveMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MatchStarted struct {
	Char rune `json:"char"`
	OpponentChar rune `json:"opponentChar"`
}

type MoveRes struct {
	Approved        bool   `json:"approved"`
	Reason string `json:"reason"`
}

type WinMessage struct {
	Status string `json:"status"`
	Cause string `json:"cause"`
}

type NotAllowedErrMessage struct {
	Reason string `json:"reason"`
}

// From server
const serverMessagesCount = 5
const (
	TMatchStarted = iota
	TMoveAns
	TOpponentMove
	TWinEvent
	TNotAllowedErr
)

// From client
type ClientMsg int
const (
	Move ClientMsg = iota
)

func (msgT ClientMsg) String() string  { 
	switch msgT {
	case Move:
		return "move"
	default:
		assert.Never("unknown type of client message", "client message", msgT)
		return "unknown"
	}
}

func UnmarshalMessage(bytes []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(bytes, &msg)

	if err != nil {
		return msg, errors.New("can't unmarshal message")
	}
	
	return msg, nil
}

func (msg Message) MarshallMessage() []byte {
	bytes, err := json.Marshal(&msg)

	assert.NoError(err, "cannot marshall message: json marshal returned error")
	return bytes
}

func MakeMessage[T any](msgType int, msgData *T) Message {
	assert.Assert(msgType < serverMessagesCount && msgType >= 0, "msgType was out of range")

	data, err := json.Marshal(&msgData)
	assert.NoError(err, "cannot marshall message: json marshal returned error")

	return Message{
		Type: msgType,
		Data: data,
	} 
}

func ParseMessage[T any](msg Message) (T, error) {
	var result T
	err := json.Unmarshal(msg.Data, &result)

	if err != nil {
		return result, errors.New(fmt.Sprint("can't parse message to type: ", reflect.TypeOf(result)))
	}

	return result, nil
}