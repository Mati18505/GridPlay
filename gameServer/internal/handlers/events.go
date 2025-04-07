package handlers

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"
	"TicTacToe/gameServer/message/client"
	"errors"

	"github.com/google/uuid"
)

type EventDisconnect struct {
	ConnectionId uuid.UUID
	Player *Player
}

type EventRemoveRoom struct {
	RoomUUID uuid.UUID
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

func (eType EventDisconnect) GetType() event.EventType {
	return event.EventTypeDisconnect;
}
func (eType EventRemoveRoom) GetType() event.EventType {
	return event.EventTypeRemoveRoom;
}
func (eType EventMove) GetType() event.EventType {
	return event.EventTypeMove;
}
func (eType EventSendMessage) GetType() event.EventType {
	return event.EventTypeSendMessage;
}

func EventFromClientMessage(msg message.Message) (event.Event, error) {
	assert.NotNil(msg, "message was nil")

	switch client.MsgType(msg.Type) {
	case client.TMove:
		moveMsg, err := message.GetConcreteMessage[client.MoveMessage](msg)
		if err != nil {
			return nil, err
		}

		return EventMove{
			X: moveMsg.X,
			Y: moveMsg.Y,
		}, nil

	default:
		return nil, errors.New("this message has no corresponding event")
	}
}