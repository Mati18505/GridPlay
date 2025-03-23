package gameServer

import (
	"errors"
)

type Event struct {
	eType EventType
}

func (e *Event) GetType() EventType {
	return e.eType
}

func CreateEvent(eType EventType) Event {
	return Event{
		eType: eType,
	}
}

type EventType interface{}

type EventExit struct {
	EventType
}

type EventMove struct {
	EventType
	x int
	y int
}

func eventTypeFromMessage(msg *message) (EventType, error) {
	switch ClientMsg(msg.Type) {
	case Move:
		data, err := ParseMessage[moveMessage](msg)

		if err != nil {
			return nil, err
		}

		return EventMove{
			x: data.X,
			y: data.Y,
		}, nil

	default:
		return nil, errors.New("This message has no corresponding event.")
	}
}