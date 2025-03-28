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
	assert.NotNil(pConnections, "player connection array was nil")

	room := &Room{
		nextHandler: nextHandler,
		uuid: uuid,
	}
	room.sync = CreateSynchronizer(room)
	room.players = room.createPlayers(pConnections)
	room.game = room.createGame()

	assert.NotNil(room.sync, "room sync was nil")
	assert.NotNil(room.game, "game was nil")
	assert.NotNil(room.players, "players was nil")

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
	pConn.SetNextHandler(&player)

	return &player
}

func (room *Room) createGame() *game.Game {
	game := game.CreateGame()
	assert.NotNil(game, "game was nil")

	return game
}

func (room *Room) sendMatchStartedMessage(player *Player) {
	assert.NotNil(player, "player was nil")
	assert.NotNil(room.game, "game was nil")
	assert.NotNil(room.nextHandler, "room next handler was nil")

	gamePlayer := room.game.GetPlayerWithId(player.playerID)
	playerChar := gamePlayer.GetChar()
	opponentChar := game.OpponentChar(playerChar)

	matchStartMsg, err := message.MakeMessage(message.TMatchStarted, &message.MatchStarted{
		Char: playerChar.GetRune(),
		OpponentChar: opponentChar.GetRune(),
	})

	if err != nil {
		assert.Never("cannot make match started message")
	}

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: *matchStartMsg,
	})
}

func (room *Room) Update() {
	assert.NotNil(room.sync, "room sync was nil")

	room.sync.SyncTransferAll(); 
}

func (room *Room) Handle(e event.Event) { 
	assert.NotNil(room.nextHandler, "room next handler was nil")

	eType := e.GetType()

	log.Printf("event in room: Type: %v, ", eType)

	switch eType {
	case event.EventTypeMove:
		eMove, _ := e.(EventMove)
		assert.NotNil(eMove.Player, "event move player was nil")

		err := room.handleMove(eMove)

		if err != nil {
			fmt.Printf("Room handle move error: %s", err)
		}
		err = room.checkGameWin(eMove)
		assert.NoError(err, "check game win error")

	case event.EventTypeExit:
		eExit, _ := e.(EventExit)
		assert.NotNil(eExit.Player, "event exit player was nil")

		opponent := room.GetOpponent(eExit.Player.playerID)

		eExit.OpponentConnId = opponent.connectionID
		eExit.RoomUUID = room.GetUUID()

		room.handleExit(eExit)

		room.nextHandler.Handle(eExit)

	default:
		room.nextHandler.Handle(e)
	}
}

func (room *Room) handleExit(eExit EventExit) {
	assert.NotNil(room.game, "game was nil")

	if room.game.GetWinState() == winState.Values.None {
		err := room.gameEndWinHandler(eExit.OpponentConnId, eExit.ConnectionId)
		assert.NoError(err, "game win handler error")
	}
}

func (room *Room) handleMove(eMove EventMove) error {
	assert.NotNil(eMove.Player, "event move player was nil")

	err := room.eMovePlayer(eMove)
	sendErr := room.eMoveSendResponse(err, eMove.Player)

	if err != nil {
		return err
	} 

	if sendErr != nil {
		return sendErr
	}
	
	opponent := room.GetOpponent(eMove.Player.playerID)

	err = room.eMoveSendMessageToOpponent(eMove, opponent)
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

func (room *Room) eMoveSendResponse(err error, player *Player) error {
	assert.NotNil(player, "player was nil")
	assert.NotNil(room.nextHandler, "room next handler was nil")

	response := new(message.MoveRes) 

	if err != nil {
		response.Approved = false
		response.Reason = err.Error()
		log.Printf("cannot handle move for %+v\n%s", player, err)
	} else {
		response.Approved = true
	}

	resMsg, err := message.MakeMessage(int(message.TMoveAns), response) 
	if err != nil {
		return err
	}

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: *resMsg,
	})

	return err
}

func (room *Room) eMoveSendMessageToOpponent(eMove EventMove, opponent *Player) error {
	assert.NotNil(opponent, "opponent was nil")
	assert.NotNil(room.nextHandler, "room next handler was nil")

	msgForOpponent, err := message.MakeMessage(message.TOpponentMove, &message.MoveMessage{
		X: eMove.X,
		Y: eMove.Y,
	})
	if err != nil {
		return err
	}

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: opponent.connectionID,
		Msg: *msgForOpponent,
	})
	return err
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
	assert.NotNil(room.nextHandler, "room next handler was nil")

	winMsg, err := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "win",
		Cause: "",
	})
	if err != nil {
		return err
	} 

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: winner,
		Msg: *winMsg,
	})
	
	loseMsg, err := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "lose",
		Cause: "",
	})

	if err != nil {
		return err
	} 

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: loser,
		Msg: *loseMsg,
	})

	return nil
}

func (room *Room) gameEndDrawHandler(c1, c2 uuid.UUID) error {
	assert.NotNil(room.nextHandler, "room next handler was nil")

	drawMsg, err := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "draw",
		Cause: "",
	})
	if err != nil {
		return err
	} 

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: c1,
		Msg: *drawMsg,
	})

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: c2,
		Msg: *drawMsg,
	})

	return nil
}
