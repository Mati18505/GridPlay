package handlers

import (
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"
	"errors"

	"github.com/google/uuid"
)

type EventExit struct {
	RoomUUID uuid.UUID
	Player *Player
	ConnectionId uuid.UUID
	OpponentConnId uuid.UUID
}

type EventMove struct {
	X int
	Y int
	Player *Player
}

type EventSendMessage struct {
	ConnectionId uuid.UUID
	Msg message.Message
}

func (eType EventExit) GetEventType() event.EventTypeEnum {
	return event.EventTypeExit;
}
func (eType EventMove) GetEventType() event.EventTypeEnum {
	return event.EventTypeMove;
}
func (eType EventSendMessage) GetEventType() event.EventTypeEnum {
	return event.EventTypeSendMessage;
}

func EventTypeFromMessage(msg *message.Message) (event.EventType, error) {
	switch message.ClientMsg(msg.Type) {
	case message.Move:
		data, err := message.ParseMessage[message.MoveMessage](msg)

		if err != nil {
			return nil, err
		}

		return EventMove{
			X: data.X,
			Y: data.Y,
		}, nil

	default:
		return nil, errors.New("this message has no corresponding event")
	}
}