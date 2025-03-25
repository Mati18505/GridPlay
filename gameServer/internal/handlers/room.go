package handlers

import (
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
	game        *game.Game
	players [2]*Player
}

func CreateRoom(nextHandler Handler, pConnections [2]*PlayerConnection, uuid uuid.UUID) *Room {
	room := &Room{
		nextHandler: nextHandler,
		uuid: uuid,
	}

	room.createPlayers(pConnections)
	room.createGame()

	return room
}

func (room *Room) GetUUID() uuid.UUID {
	return room.uuid
}

func (room *Room) createPlayers(pConnections [2]*PlayerConnection) {
	room.players[0] = room.createPlayer(pConnections[0], 0)
	room.players[1] = room.createPlayer(pConnections[1], 1)
}

func (room *Room) createPlayer(pConn *PlayerConnection, playerId int) *Player {
	player := CreatePlayer(room, pConn.id, playerId)
	pConn.SetNextHandler(&player)

	return &player
}

func (room *Room) createGame() {
	room.game = game.CreateGame()

	room.sendMatchStartedMessage(room.players[0])
	room.sendMatchStartedMessage(room.players[1])
}

func (room *Room) sendMatchStartedMessage(player *Player) {
	gamePlayer := room.game.GetPlayerWithId(player.playerID)
	playerChar := gamePlayer.GetChar()
	opponentChar := game.OpponentChar(playerChar)

	matchStartMsg, err := message.MakeMessage(message.TMatchStarted, &message.MatchStarted{
		Char: rune(playerChar),
		OpponentChar: rune(opponentChar),
	})

	if err != nil {
		log.Print("cannot make match started message")
		// TODO: what now?
	}

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: *matchStartMsg,
	})
}

func (room *Room) Handle(e event.Event) { 
	eType := e.GetType()

	log.Printf("event in room: Type: %v, ", eType)

	switch eType {
	case event.EventTypeMove:
		eMove, _ := e.(EventMove)

		err := room.handleMove(eMove)

		if err != nil {
			fmt.Printf("Room handle move error: %s", err)
		}
		room.checkGameWin(eMove)

	case event.EventTypeExit:
		eExit, _ := e.(EventExit)
		opponent := room.GetOpponent(eExit.Player.playerID)

		eExit.RoomUUID = room.GetUUID()
		eExit.OpponentConnId = opponent.connectionID

		room.handleExit(eExit)

		room.nextHandler.Handle(eExit)

	default:
		room.nextHandler.Handle(e)
	}
}

func (room *Room) handleExit(eExit EventExit) {
	if room.game.GetWinState() == winState.Values.None {
		room.gameEndWinHandler(eExit.OpponentConnId, eExit.ConnectionId)
	}
}

func (room *Room) handleMove(eMove EventMove) error {
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
	// TODO: assert if game don't exist ? (maybe always exist)
	var err error
	currPlayer := room.game.GetCurrentRoundPlayer()

	gamePlayer := room.game.GetPlayerWithId(eMove.Player.playerID)

	if err != nil {
		return err
	}

	if currPlayer == gamePlayer {
		err = room.game.Move(game.Pos{X: eMove.X, Y: eMove.Y})
	} else {
		err = errors.New("not your round, dummy")
	}

	return err
}

func (room *Room) eMoveSendResponse(err error, player *Player) error {
	// TODO: assert player, err != nil

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
	// TODO: assert if game don't exist ? (maybe always exist)
	var opponentId int

	switch playerID {
	case 0:
		opponentId = 1
	case 1:
		opponentId = 0
	}

	return room.players[opponentId]
}

func (room *Room) gameEndWinHandler(winner, loser uuid.UUID) error {
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

	if err != nil {
		return err
	}
	
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

	if err != nil {
		return err
	}

	return nil
}

func (room *Room) gameEndDrawHandler(c1, c2 uuid.UUID) error {
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

	if err != nil {
		return err
	}

	room.nextHandler.Handle(EventSendMessage{
		ConnectionId: c2,
		Msg: *drawMsg,
	})

	if err != nil {
		return err
	}

	return nil
}
