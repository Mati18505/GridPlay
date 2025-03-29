package handlers

import (
	"TicTacToe/assert"
	"TicTacToe/game"
	"TicTacToe/game/winState"
	"errors"
	"fmt"
	"log"

	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"

	"github.com/google/uuid"
)

type Room struct {
	nextHandler Handler
	uuid uuid.UUID
	sync *Synchronizer
	game        *game.Game
	players [2]*Player
}

func CreateRoom(nextHandler Handler, pConnections [2]*PlayerConnection, uuid uuid.UUID) *Room {
	assert.NotNil(nextHandler, "next handler was nil")
	assert.NotNil(pConnections[0], "player connection was nil")
	assert.NotNil(pConnections[1], "player connection was nil")

	room := &Room{
		nextHandler: nextHandler,
		uuid: uuid,
	}
	room.sync = CreateSynchronizer(room)
	room.players = room.createPlayers(pConnections)
	room.game = room.createGame()

	assert.NotNil(room.sync, "room sync was nil")
	assert.NotNil(room.game, "game was nil")

	room.sendMatchStartedMessage(room.players[0])
	room.sendMatchStartedMessage(room.players[1])

	return room
}

func (room *Room) GetUUID() uuid.UUID {
	return room.uuid
}

func (room *Room) createPlayers(pConnections [2]*PlayerConnection) [2]*Player {
	assert.NotNil(pConnections, "player connection array was nil")

	var players [2]*Player
	players[0] = room.createPlayer(pConnections[0], 0)
	players[1] = room.createPlayer(pConnections[1], 1)
	assert.NotNil(players[0], "player was nil")
	assert.NotNil(players[1], "player was nil")

	return players
}

func (room *Room) createPlayer(pConn *PlayerConnection, playerId int) *Player {
	assert.NotNil(room.sync, "room sync was nil")
	assert.NotNil(pConn, "player connection was nil")

	player := CreatePlayer(room.sync, pConn.uuid, playerId)
	pConn.SetNextHandler(player)

	return player
}

func (room *Room) createGame() *game.Game {
	game := game.CreateGame()
	assert.NotNil(game, "game was nil")

	return game
}

func (room *Room) sendMatchStartedMessage(player *Player) {
	assert.NotNil(player, "player was nil")
	assert.NotNil(room.game, "game was nil")

	gamePlayer := room.game.GetPlayerWithId(player.playerID)
	playerChar := gamePlayer.GetChar()
	opponentChar := game.OpponentChar(playerChar)

	matchStartMsg := message.MakeMessage(message.TMatchStarted, &message.MatchStarted{
		Char: playerChar.GetRune(),
		OpponentChar: opponentChar.GetRune(),
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: matchStartMsg,
	})
}

func (room *Room) Update() {
	assert.NotNil(room.sync, "room sync was nil")

	room.sync.SyncTransferAll(); 
}

func (room *Room) Handle(e event.Event) { 
	eType := e.GetType()

	log.Printf("event in room: Type: %v, ", eType)

	switch eType {
	case event.EventTypeMove:
		eMove, ok := e.(EventMove)
		assert.Assert(ok, "type assertion failed for event move")

		err := room.handleMove(eMove)

		if err != nil {
			fmt.Printf("Room handle move error: %s", err)
		}

	case event.EventTypeDisconnect:
		eExit, ok := e.(EventDisconnect)
		assert.Assert(ok, "type assertion failed for event exit")

		room.handleExit(eExit)

	default:
		room.sendToNextHandler(e)
	}
}

func (room *Room) handleExit(eExit EventDisconnect) {
	assert.NotNil(eExit.Player, "event exit player was nil")
	assert.NotNil(room.game, "game was nil")

	opponent := room.GetOpponent(eExit.Player.playerID)

	eRemoveRoom := EventRemoveRoom{
		RoomUUID: room.GetUUID(),
		ConnectionId: eExit.ConnectionId,
		OpponentConnId: opponent.connectionID,
	}

	if room.game.GetWinState() == winState.Values.None {
		err := room.gameEndWinHandler(opponent.connectionID, eExit.ConnectionId)
		assert.NoError(err, "game win handler error")
	}

	room.sendToNextHandler(eRemoveRoom)
}

func (room *Room) handleMove(eMove EventMove) error {
	assert.NotNil(eMove.Player, "event move player was nil")

	err := room.eMovePlayer(eMove)

	if err != nil {
		room.eMoveSendErrorResponse(err, eMove.Player)
		return err
	} else {
		room.eMoveSendSuccessResponse(eMove.Player)
	}

	opponent := room.GetOpponent(eMove.Player.playerID)
	room.eMoveSendMessageToOpponent(eMove, opponent)

	err = room.checkGameWin(eMove)

	return err
}

func (room *Room) eMovePlayer(eMove EventMove) error {
	assert.NotNil(room.game, "game was nil")
	assert.NotNil(eMove.Player, "event move player was nil")

	var err error
	currPlayer := room.game.GetCurrentRoundPlayer()
	gamePlayer := room.game.GetPlayerWithId(eMove.Player.playerID)

	if currPlayer == gamePlayer {
		err = room.game.Move(game.Pos{X: eMove.X, Y: eMove.Y})
	} else {
		err = errors.New("not your round, dummy")
	}

	return err
}

func (room *Room) eMoveSendErrorResponse(err error, player *Player) error {
	assert.NotNil(player, "player was nil")
	assert.NotNil(err, "error was nil")

	response := new(message.MoveRes) 

	response.Approved = false
	response.Reason = err.Error()
	log.Printf("cannot handle move for %+v\n%s", player, err)

	resMsg := message.MakeMessage(int(message.TMoveAns), response) 

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: resMsg,
	})

	return err
}

func (room *Room) eMoveSendSuccessResponse(player *Player) {
	assert.NotNil(player, "player was nil")

	response := new(message.MoveRes) 
	response.Approved = true

	resMsg := message.MakeMessage(int(message.TMoveAns), response) 

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: resMsg,
	})
}

func (room *Room) eMoveSendMessageToOpponent(eMove EventMove, opponent *Player) {
	assert.NotNil(opponent, "opponent was nil")

	msgForOpponent := message.MakeMessage(message.TOpponentMove, &message.MoveMessage{
		X: eMove.X,
		Y: eMove.Y,
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: opponent.connectionID,
		Msg: msgForOpponent,
	})
}

func (room *Room) checkGameWin(eMove EventMove) error {
	assert.NotNil(room.game, "game was nil")
	assert.NotNil(eMove.Player, "event move player was nil")

	wState := room.game.GetWinState()
	player := eMove.Player
	opponent := room.GetOpponent(player.playerID)
	
	var err error

	if wState == winState.Values.Win {
		err = room.gameEndWinHandler(player.connectionID, opponent.connectionID)
	} else if wState == winState.Values.Draw {
		err = room.gameEndDrawHandler(player.connectionID, opponent.connectionID)
	}

	return err
}

// TODO: unit test
func (room *Room) GetOpponent(playerID int) *Player {
	assert.NotNil(room.players, "players array was nil")

	var opponentId int

	switch playerID {
	case 0:
		opponentId = 1
	case 1:
		opponentId = 0
	default:
		assert.Never("player id was out of range")
	}

	opponent := room.players[opponentId]
	assert.NotNil(opponent, "opponent was nil")

	return opponent
}

func (room *Room) gameEndWinHandler(winner, loser uuid.UUID) error {
	winMsg := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "win",
		Cause: "",
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: winner,
		Msg: winMsg,
	})
	
	loseMsg := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "lose",
		Cause: "",
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: loser,
		Msg: loseMsg,
	})

	return nil
}

func (room *Room) gameEndDrawHandler(c1, c2 uuid.UUID) error {
	drawMsg := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "draw",
		Cause: "",
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: c1,
		Msg: drawMsg,
	})

	room.sendToNextHandler(EventSendMessage{
		ConnectionId: c2,
		Msg: drawMsg,
	})

	return nil
}

func (room *Room) sendToNextHandler(e event.Event) {
	assert.NotNil(room.nextHandler, "room next handler was nil")

	room.nextHandler.Handle(e)
}