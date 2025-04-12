package externalEvent

import (
	"GridPlay/assert"
)


type Event interface {
	GetType() EventType
}

type EmptyEvent struct {}

func (emptyEvent EmptyEvent) GetType() EventType {
	return EventTypeNone;
}

type EventType int
const (
	EventTypeNone EventType = iota
	EventTypeGameMessage
)

func (eType EventType) String() string {
	switch eType {
	case EventTypeNone:
		return "None"
	case EventTypeGameMessage:
		return "GameMessage"
	default:
		assert.Never("unknown type of event", "event", eType)
		return "Unknown"
	}
}

func (eType EventGameMessage) GetType() EventType {
	return EventTypeGameMessage;
}

type EventGameMessage struct {
	Name string
	Data any
	PId int
}