package room

import (
	"GridPlay/assert"
	"GridPlay/gameServer/internal/handlers"
	"errors"
)

type gameEnded struct {
	room *Room
}

func createGameEnded(room *Room) *gameEnded {
	assert.NotNil(room, "room was nil")

	return &gameEnded{
		room: room,
	};
}

func (state *gameEnded) handleDisconnect(playerId, opponentId int) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	eRemoveRoom := handlers.EventRemoveRoom{
		RoomUUID: room.GetUUID(),
	}

	room.players[playerId] = nil
	room.sendToNextHandler(eRemoveRoom)
}

func (state *gameEnded) handleMove(eMove handlers.EventMove) {
	state.room.eMoveSendErrorResponse(errors.New("cannot move because game has ended"), eMove.Player)
}