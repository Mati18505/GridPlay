package handlers

import (
	"log/slog"

	"github.com/google/uuid"

	"GridPlay/assert"
	"GridPlay/gameServer/internal/event"
)

type Player struct {
	nextHandler Handler
	connectionID uuid.UUID
	playerID int
}

func CreatePlayer(nextHandler Handler, connId uuid.UUID, playerId int) *Player {
	assert.NotNil(nextHandler, "nextHandler was nil")

	if playerId < 0 || playerId > 1 {
		assert.Never("player id was out of range")
	}

	return &Player{
		nextHandler: nextHandler,
		connectionID: connId,
		playerID: playerId,
	}
}

func (player *Player) Handle(e event.Event) {
	eType := e.GetType()

	slog.Debug("event in player", "Type", eType, "event", e)

	switch eType {
	case event.EventTypeMove:
		eMove, ok := e.(EventMove)
		assert.Assert(ok, "type assertion failed for event move")

		player.handleMove(eMove)

	case event.EventTypeDisconnect:
		eDisconnect, ok := e.(EventDisconnect)
		assert.Assert(ok, "type assertion failed for event Disconnect")

		player.handleDisconnect(eDisconnect)

	default:
		player.sendToNextHandler(e)
	}
}

func (player *Player) handleMove(eMove EventMove) {
	eMove.Player = player
	player.sendToNextHandler(eMove)
}

func (player *Player) handleDisconnect(eDisconnect EventDisconnect) {
	eDisconnect.Player = player
	player.sendToNextHandler(eDisconnect)
}

func (player *Player) sendToNextHandler(e event.Event) {
	assert.NotNil(player.nextHandler, "player next handler was nil")

	player.nextHandler.Handle(e)
}