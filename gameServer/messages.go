package gameServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type message struct {
	Type int `json:"type"`
	Data json.RawMessage `json:"data"`
}

type moveMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type matchStarted struct {
	Char rune `json:"char"`
	OpponentChar rune `json:"opponentChar"`
}

type moveRes struct {
	Approved        bool   `json:"approved"`
	Reason string `json:"reason"`
}

type winMessage struct {
	Status string `json:"status"`
	Cause string `json:"cause"`
}

// From server
const (
	MatchStarted = iota
	MoveAns
	OpponentMove
	WinEvent
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

func UnmarshalMessage(bytes []byte) (*message, error) {
	msg := new(message)

	err := json.Unmarshal(bytes, &msg)
	if err != nil {
		return nil, errors.New("can't unmarshal message")
	}
	
	return msg, nil
}

func MarshallMessage(msg *message) ([]byte, error) {
	bytes, err := json.Marshal(&msg)
	if err != nil {
		return nil, errors.New("can't marshall message")
	}
	return bytes, nil
}

func MakeMessage[T any](msgType int, msgData *T) (*message, error) {
	data, err := json.Marshal(msgData)
	if err != nil {
		return nil, err
	}
	return &message{
		Type: msgType,
		Data: data,
	}, nil
}

func ParseMessage[T any](msg *message) (*T, error) {
	result := new(T)

	err := json.Unmarshal(msg.Data, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprint("can't parse message to type: ", reflect.TypeOf(result)))
	}

	return result, nil
}