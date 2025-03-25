package handlers

import (
	"log"

	"github.com/google/uuid"

	"TicTacToe/gameServer/internal/event"
)

type Player struct {
	nextHandler Handler
	connectionID uuid.UUID
	playerID int
	playerChan chan event.Event
}

func CreatePlayer(nextHandler Handler, connId uuid.UUID, playerId int) Player {
	return Player{
		nextHandler: nextHandler,
		connectionID: connId,
		playerID: playerId,
		playerChan: make(chan event.Event),
	}
}

func (player *Player) Handle(e event.Event) {
	eType := e.GetType()

	log.Printf("event in player: Type: %v, ", eType)

	switch eType.GetEventType() {
	case event.EventTypeMove:
		eMove, _ := eType.(EventMove)

		eMove.Player = player
		e := event.CreateEvent(eMove)

		player.nextHandler.Handle(e)

	case event.EventTypeExit:
		eExit, _ := eType.(EventExit)

		eExit.Player = player
		e := event.CreateEvent(eExit)

		player.nextHandler.Handle(e)

	default:
		player.nextHandler.Handle(e)
	}
}