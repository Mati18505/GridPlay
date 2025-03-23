package gameServer

import (
	"errors"

	"github.com/google/uuid"
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

type EventTypeEnum int
const (
	EventTypeExit EventTypeEnum = iota
	EventTypeMove
	EventTypeSendMessage
)

type EventType interface{
	GetEventType() EventTypeEnum
}

type EventExit struct {
	roomUUID uuid.UUID
	player *player
	connectionId uuid.UUID
	opponentConnId uuid.UUID
}

type EventMove struct {
	x int
	y int
	player *player
}

type EventSendMessage struct {
	connectionId uuid.UUID
	msg message
}

func (eType EventExit) GetEventType() EventTypeEnum {
	return EventTypeExit;
}
func (eType EventMove) GetEventType() EventTypeEnum {
	return EventTypeMove;
}
func (eType EventSendMessage) GetEventType() EventTypeEnum {
	return EventTypeSendMessage;
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
		return nil, errors.New("this message has no corresponding event")
	}
}