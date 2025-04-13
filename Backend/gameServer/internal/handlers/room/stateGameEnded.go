package room

import (
	"GridPlay/assert"
	"GridPlay/gameServer/internal/handlers"
	"GridPlay/gameServer/message/serverMsg"
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

func (state *gameEnded) handleGameMsg(eGameMsg handlers.EventGameMessage) error {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	approveMsg := serverMsg.MakeMessage(serverMsg.TApprove, &serverMsg.Approve{
		Approved: false,
		Reason: "cannot handle game_message because game has ended",
	})

	eApprove := handlers.EventSendMessage{
		ConnectionId: room.GetPlayerByGameId(eGameMsg.PlayerId).GetConnectionId(),
		Msg: approveMsg,
	}

	room.sendToNextHandler(eApprove)

	return nil
}