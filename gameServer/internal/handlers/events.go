package handlers

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message/client"
	"TicTacToe/gameServer/message/server"
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
	Msg server.Message
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

func EventFromClientMessage(msg client.Message) (event.Event, error) {
	assert.NotNil(msg, "message was nil")

	switch msg.Type {
	case client.TMove:
		moveMsg, err := client.GetConcreteMessage[client.MoveMessage](msg)
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