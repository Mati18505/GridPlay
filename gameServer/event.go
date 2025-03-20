package gameServer

import (
	"errors"
)

type Event struct {
	receiver Receiver
	eType EventType
}

func (e *Event) GetReceiver() Receiver {
	return e.receiver
}

func (e *Event) GetType() EventType {
	return e.eType
}

type Receiver int
const (
	RCV_Connection Receiver = iota
	RCV_Player
	RCV_Room
	RCV_Server
)

func CreateEvent(receiver Receiver, eType EventType) Event {
	return Event{
		receiver: receiver,
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