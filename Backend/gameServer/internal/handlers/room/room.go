package room

import (
	"GridPlay/assert"
	"GridPlay/game"
	"GridPlay/game/winState"
	"GridPlay/gameServer/message/serverMsg"
	"errors"
	"log/slog"

	"GridPlay/gameServer/internal/event"
	"GridPlay/gameServer/internal/handlers"

	"github.com/google/uuid"
)

type Room struct {
	nextHandler handlers.Handler
	uuid uuid.UUID
	sync *handlers.Synchronizer
	game        *game.Game
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

func (room *Room) createGame() *game.Game {
	game := game.CreateGame()
	assert.NotNil(game, "game was nil")

	return game
}

func (room *Room) sendMatchStartedMessage(player *handlers.Player) {
	assert.NotNil(player, "player was nil")
	assert.NotNil(room.game, "game was nil")

	gamePlayer := room.game.GetPlayerWithId(player.GetPlayerId())
	playerChar := gamePlayer.GetChar()
	opponentChar := game.OpponentChar(playerChar)

	matchStartMsg := serverMsg.MakeMessage(serverMsg.TMatchStarted, &serverMsg.MatchStarted{
		Char: playerChar.GetRune(),
		OpponentChar: opponentChar.GetRune(),
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: player.GetConnectionId(),
		Msg: matchStartMsg,
	})
}

func (room *Room) Update() {
	assert.NotNil(room.sync, "room sync was nil")

	room.sync.SyncTransferAll(); 
}

func (room *Room) Handle(e event.Event) { 
	eType := e.GetType()

	slog.Debug("event in room", "Type", eType, "event", e)

	switch eType {
	case event.EventTypeMove:
		eMove, ok := e.(handlers.EventMove)
		assert.Assert(ok, "type assertion failed for event move")

		room.handleMove(eMove)

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

func (room *Room) handleMove(eMove handlers.EventMove) {
	assert.NotNil(eMove.Player, "event move player was nil")

	room.state.handleMove(eMove)
}

func (room *Room) eMovePlayer(eMove handlers.EventMove) error {
	assert.NotNil(room.game, "game was nil")
	assert.NotNil(eMove.Player, "event move player was nil")

	var err error
	currPlayer := room.game.GetCurrentRoundPlayer()
	gamePlayer := room.game.GetPlayerWithId(eMove.Player.GetPlayerId())

	if currPlayer == gamePlayer {
		err = room.game.Move(game.Pos{X: eMove.X, Y: eMove.Y})
	} else {
		err = errors.New("not your round, dummy")
	}

	return err
}

func (room *Room) eMoveSendErrorResponse(err error, player *handlers.Player) {
	assert.NotNil(player, "player was nil")
	assert.NotNil(err, "error was nil")

	slog.Info("cannot handle move for", "player uuid", player.GetConnectionId().String(), "player game id", player.GetPlayerId(), "err", err)

	msg := serverMsg.MakeMessage(serverMsg.TMoveAns, serverMsg.MoveRes{
		Approved: false,
		Reason: err.Error(),
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: player.GetConnectionId(),
		Msg: msg,
	})
}

func (room *Room) eMoveSendSuccessResponse(player *handlers.Player) {
	assert.NotNil(player, "player was nil")

	msg := serverMsg.MakeMessage(serverMsg.TMoveAns, serverMsg.MoveRes{
		Approved: true,
	})
	

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: player.GetConnectionId(),
		Msg: msg,
	})
}

func (room *Room) eMoveSendMessageToOpponent(eMove handlers.EventMove, opponent *handlers.Player) {
	assert.NotNil(opponent, "opponent was nil")

	msgForOpponent := serverMsg.MakeMessage(serverMsg.TOpponentMove, &serverMsg.MoveMessage{
		X: eMove.X,
		Y: eMove.Y,
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: opponent.GetConnectionId(),
		Msg: msgForOpponent,
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
	
	winMsg := serverMsg.MakeMessage(serverMsg.TWinEvent, &serverMsg.WinMessage{
		Status: "win",
		Cause: "",
	})

	room.sendToNextHandler(handlers.EventSendMessage{
		ConnectionId: winner,
		Msg: winMsg,
	})
	
	loseMsg := serverMsg.MakeMessage(serverMsg.TWinEvent, &serverMsg.WinMessage{
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
	
	winMsg := serverMsg.MakeMessage(serverMsg.TWinEvent, &serverMsg.WinMessage{
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

	drawMsg := serverMsg.MakeMessage(serverMsg.TWinEvent, &serverMsg.WinMessage{
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

	return room.game.GetWinState() != winState.Values.None
}