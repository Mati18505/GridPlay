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
}

func CreatePlayer(nextHandler Handler, connId uuid.UUID, playerId int) Player {
	return Player{
		nextHandler: nextHandler,
		connectionID: connId,
		playerID: playerId,
	}
}

func (player *Player) Handle(e event.Event) {
	eType := e.GetType()

	log.Printf("event in player: Type: %v, ", eType)

	switch eType {
	case event.EventTypeMove:
		player.handleMove(e.(EventMove))

	case event.EventTypeExit:
		player.handleExit(e.(EventExit))

	default:
		player.nextHandler.Handle(e)
	}
}

func (player *Player) handleMove(eMove EventMove){
	eMove.Player = player
	player.nextHandler.Handle(eMove)
}

func (player *Player) handleExit(eExit EventExit){
	eExit.Player = player
	player.nextHandler.Handle(eExit)
}