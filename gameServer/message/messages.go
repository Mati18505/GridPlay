package message

import (
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
		return "unknown"
	}
}

func UnmarshalMessage(bytes []byte) (*Message, error) {
	msg := new(Message)

	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, errors.New("can't unmarshal message")
	}
	
	return msg, nil
}

func MarshallMessage(msg *Message) ([]byte, error) {
	bytes, err := json.Marshal(&msg)
	if err != nil {
		return nil, errors.New("can't marshall message")
	}
	return bytes, nil
}

func MakeMessage[T any](msgType int, msgData *T) (*Message, error) {
	data, err := json.Marshal(msgData)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type: msgType,
		Data: data,
	}, nil
}

func ParseMessage[T any](msg *Message) (*T, error) {
	result := new(T)

	err := json.Unmarshal(msg.Data, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprint("can't parse message to type: ", reflect.TypeOf(result)))
	}

	return result, nil
}