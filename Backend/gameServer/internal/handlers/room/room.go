package room

import (
	"GridPlay/assert"
	"GridPlay/game"
	"GridPlay/game/chessGame"
	"GridPlay/gameServer/message/serverMsg"
	"log/slog"

	"GridPlay/gameServer/internal/event"
	"GridPlay/gameServer/internal/handlers"

	"github.com/google/uuid"
)

type Room struct {
	nextHandler handlers.Handler
	uuid uuid.UUID
	sync *handlers.Synchronizer
	game        game.Game
	players [2]*handlers.Player
	state state
}

func CreateRoom(nextHandler handlers.Handler, pConnections [2]*handlers.PlayerConnection, uuid uuid.UUID) *Room {
	assert.NotNil(nextHandler, "next handler was nil")
	assert.NotNil(pConnections[0], "player connection was nil")
	assert.NotNil(pConnections[1], "player connection was nil")

	room := &Room{
		nextHandler: nextHandler,
		uuid: uuid,
		state: nil,
	}
	room.sync = handlers.CreateSynchronizer(room)
	room.players = room.createPlayers(pConnections)
	room.game = room.createGame()

	assert.NotNil(room.sync, "room sync was nil")
	assert.NotNil(room.game, "game was nil")
	room.startGame()

	return room
}

func (room *Room) setState(state state) {
	room.state = state
}

func (room *Room) startGame() {
	room.setState(createGameActive(room))
}

func (room *Room) GetUUID() uuid.UUID {
	return room.uuid
}

func (room *Room) createPlayers(pConnections [2]*handlers.PlayerConnection) [2]*handlers.Player {
	assert.NotNil(pConnections, "player connection array was nil")

	var players [2]*handlers.Player
	players[0] = room.createPlayer(pConnections[0], 0)
	players[1] = room.createPlayer(pConnections[1], 1)
	assert.NotNil(players[0], "player was nil")
	assert.NotNil(players[1], "player was nil")

	return players
}

func (room *Room) createPlayer(pConn *handlers.PlayerConnection, playerId int) *handlers.Player {
	assert.NotNil(room.sync, "room sync was nil")
	assert.NotNil(pConn, "player connection was nil")

	player := handlers.CreatePlayer(room.sync, pConn.GetUUID(), playerId)
	pConn.SetNextHandler(player)

	return player
}

func (room *Room) createGame() game.Game {
	game := chessGame.CreateChessGame()
	assert.NotNil(game, "game was nil")

	return game
}

func (room *Room) Update() {
	assert.NotNil(room.sync, "room sync was nil")

	room.sync.SyncTransferAll(); 
}

func (room *Room) Handle(e event.Event) { 
	eType := e.GetType()

	slog.Debug("event in room", "Type", eType, "event", e)

	switch eType {
	case event.EventTypeGameMessage:
		eGameMsg, ok := e.(handlers.EventGameMessage)
		assert.Assert(ok, "type assertion failed for event move")

		room.handleGameMsg(eGameMsg)

	case event.EventTypeDisconnect:
		eDisconnect, ok := e.(handlers.EventDisconnect)
		assert.Assert(ok, "type assertion failed for event disconnect")

		room.handleDisconnect(eDisconnect)

	default:
		room.sendToNextHandler(e)
	}
}

func (room *Room) handleDisconnect(eDisconnect handlers.EventDisconnect) {
	assert.NotNil(eDisconnect.Player, "event disconnect player was nil")
	assert.NotNil(room.game, "game was nil")

	room.sendToNextHandler(eDisconnect)

	playerId := eDisconnect.Player.GetPlayerId()
	opponentId := room.GetOpponentId(playerId)

	room.state.handleDisconnect(playerId, opponentId)
}

func (room *Room) handleGameMsg(eGameMsg handlers.EventGameMessage) {
	assert.NotNil(eGameMsg.Player, "event move player was nil")

	err := room.state.handleGameMsg(eGameMsg)
	if err != nil {
		slog.Error("Cannot handle game message", "err", err.Error())
	}
}

func (room *Room) sendGameAnswer(player *handlers.Player, eGameMsg handlers.EventGameMessage) {
	assert.NotNil(player, "player was nil")

	msg := serverMsg.MakeMessage(serverMsg.TGameMessage, serverMsg.GameMessage{
		Data: eGameMsg.Data,
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: player.GetConnectionId(),
		Msg: msg,
	})
}

// TODO: unit test
func (room *Room) GetOpponentId(playerID int) int {
	var opponentId int

	switch playerID {
	case 0:
		opponentId = 1
	case 1:
		opponentId = 0
	default:
		assert.Never("player id was out of range")
	}

	return opponentId
}


func (room *Room) gameEndWinHandler(winner, loser uuid.UUID) {
	slog.Debug("game win", "room", room.uuid, "winner", winner)
	
	winMsg := serverMsg.MakeMessage(serverMsg.TGameEnded, &serverMsg.GameEnded{
		Status: "win",
		Cause: "",
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: winner,
		Msg: winMsg,
	})
	
	loseMsg := serverMsg.MakeMessage(serverMsg.TGameEnded, &serverMsg.GameEnded{
		Status: "lose",
		Cause: "",
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: loser,
		Msg: loseMsg,
	})
}

func (room *Room) gameEndWinOnePlayerHandler(winner uuid.UUID) {
	slog.Debug("game win", "room", room.uuid, "winner", winner)
	
	winMsg := serverMsg.MakeMessage(serverMsg.TGameEnded, &serverMsg.GameEnded{
		Status: "win",
		Cause: "",
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: winner,
		Msg: winMsg,
	})
}


func (room *Room) gameEndDrawHandler(c1, c2 uuid.UUID) {
	slog.Debug("game draw", "room", room.uuid)

	drawMsg := serverMsg.MakeMessage(serverMsg.TGameEnded, &serverMsg.GameEnded{
		Status: "draw",
		Cause: "",
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: c1,
		Msg: drawMsg,
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: c2,
		Msg: drawMsg,
	})
}

func (room *Room) sendToNextHandler(e event.Event) {
	assert.NotNil(room.nextHandler, "room next handler was nil")

	room.nextHandler.Handle(e)
}

func (room *Room) gameHasEnded() bool {
	assert.NotNil(room.game, "game was nil")

	return room.game.GetWinState().T != game.None
}