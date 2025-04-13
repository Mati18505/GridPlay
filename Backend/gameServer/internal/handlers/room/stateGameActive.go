package room

import (
	"GridPlay/assert"
	"GridPlay/game"
	"GridPlay/gameServer/externalEvent"
	"GridPlay/gameServer/internal/handlers"
)

type gameActive struct {
	room *Room
}

func createGameActive(room *Room) *gameActive {
	assert.NotNil(room, "room was nil")

	state := &gameActive{
		room: room,
	};

	state.sendGameStartMsgs()

	return state
}

func (state *gameActive) sendGameStartMsgs() {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	for i := range 2 {
		gameStartAns := room.game.GetGameStartMessage(i)
		eGameStart := state.createGameMsgFromGameAns(gameStartAns)
		room.sendGameAnswer(eGameStart)
	}
}

func (state *gameActive) handleDisconnect(playerId, opponentId int) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.players, "players was nil")

	opponent := room.GetPlayerByGameId(opponentId)

	// This player disconnect first, so opponent should be online, in room.
	assert.NotNil(opponent, "opponent should not be nil")

	if !room.gameHasEnded() {
		room.gameEndWinOnePlayerHandler(opponent.GetConnectionId())
	}

	room.players[playerId] = nil
	
	room.setState(createGameEnded(room))
}

func (state *gameActive) handleGameMsg(eGameMsg handlers.EventGameMessage) error {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	externalGameMsg := externalEvent.EventGameMessage{
		Data: eGameMsg.Data,
		PId: eGameMsg.PlayerId,
	}

	gameAnswers, err := room.game.HandleGameMsg(externalGameMsg)

	if err != nil {
		return err
	}

	for _, gameAnswer := range gameAnswers {
		eGameMsg := state.createGameMsgFromGameAns(gameAnswer)
		room.sendGameAnswer(eGameMsg)
	}

	state.checkGameWin(eGameMsg)

	return nil
}

func (state *gameActive) createGameMsgFromGameAns(gameAnswer externalEvent.EventGameMessage) handlers.EventGameMessage {
	return handlers.EventGameMessage{
		Name: gameAnswer.Name,
		Data: gameAnswer.Data,
		PlayerId: gameAnswer.PId,
	}
}

func (state *gameActive) checkGameWin(eGameMsg handlers.EventGameMessage) {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.game, "game was nil")

	wState := room.game.GetWinState()
	player := room.GetPlayerByGameId(eGameMsg.PlayerId)
	opponent := state.GetOpponent(eGameMsg.PlayerId)
	
	if wState.T == game.Win {
		room.gameEndWinHandler(player.GetConnectionId(), opponent.GetConnectionId())
	} else if wState.T == game.Draw {
		room.gameEndDrawHandler(player.GetConnectionId(), opponent.GetConnectionId())
	}
}

func (state *gameActive) GetOpponent(playerId int) *handlers.Player {
	assert.NotNil(state.room, "room was nil")
	room := state.room

	assert.NotNil(room.players, "players was nil")
	
	opponentId := room.GetOpponentId(playerId)
	opponent := room.GetPlayerByGameId(opponentId)

	// Game is active, so opponent shouldn't be nil.
	assert.NotNil(opponent, "opponent was nil")
	
	return opponent
}
