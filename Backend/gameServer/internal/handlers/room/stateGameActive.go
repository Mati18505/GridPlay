package room

import (
	"GridPlay/assert"
	"GridPlay/game/winState"
	"GridPlay/gameServer/internal/handlers"
)

type gameActive struct {
	state
	room *Room
}

func createGameActive(room *Room) *gameActive {
	assert.NotNil(room, "room was nil")

	room.sendMatchStartedMessage(room.players[0])
	room.sendMatchStartedMessage(room.players[1])

	return &gameActive{
		room: room,
	};
}

func (state *gameActive) handleDisconnect(playerId, opponentId int) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.players, "players was nil")

	opponent := room.players[opponentId]

	// This player disconnect first, so opponent should be online, in room.
	assert.NotNil(opponent, "opponent should not be nil")

	if !room.gameHasEnded() {
		room.gameEndWinOnePlayerHandler(opponent.GetConnectionId())
	}

	room.players[playerId] = nil
	
	room.setState(createGameEnded(room))
}

func (state *gameActive) handleMove(eMove handlers.EventMove) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	err := room.eMovePlayer(eMove)

	if err != nil {
		room.eMoveSendErrorResponse(err, eMove.Player)
		return
	}

	room.eMoveSendSuccessResponse(eMove.Player)

	opponent := state.GetOpponent(eMove.Player.GetPlayerId())
	room.eMoveSendMessageToOpponent(eMove, opponent)

	state.checkGameWin(eMove)
}

func (state *gameActive) checkGameWin(eMove handlers.EventMove) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.game, "game was nil")
	assert.NotNil(eMove.Player, "event move player was nil")

	wState := room.game.GetWinState()
	player := eMove.Player
	opponent := state.GetOpponent(player.GetPlayerId())
	
	if wState == winState.Values.Win {
		room.gameEndWinHandler(player.GetConnectionId(), opponent.GetConnectionId())
	} else if wState == winState.Values.Draw {
		room.gameEndDrawHandler(player.GetConnectionId(), opponent.GetConnectionId())
	}
}

func (state *gameActive) GetOpponent(playerId int) *handlers.Player {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.players, "players was nil")
	
	opponentId := room.GetOpponentId(playerId)
	opponent := room.players[opponentId]

	// Game is active, so opponent shouldn't be nil.
	assert.NotNil(opponent, "opponent was nil")
	
	return opponent
}