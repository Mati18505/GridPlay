package handlers

import (
	"TicTacToe/game"
	"TicTacToe/game/winState"
	"errors"
	"log"

	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"

	"github.com/google/uuid"
)

type Room struct {
	uuid uuid.UUID
	game        *game.Game
	players [2]Player
	serverChan chan<- event.Event
	roomChan chan event.Event
	stopLoop chan bool
}

func CreateRoom(pConnections [2]*PlayerConnection, uuid uuid.UUID, serverChan chan<- event.Event) *Room {
	room := &Room{
		uuid: uuid,
		serverChan: serverChan,
		roomChan: make(chan event.Event),
		stopLoop: make(chan bool),
	}

	room.createPlayers(pConnections)
	room.createGame()

	return room
}

func (room *Room) StartLoop() {
	go room.loop()
}

func (room *Room) EndLoop() {
	room.stopLoop <- true
}

func (room *Room) GetUUID() uuid.UUID {
	return room.uuid
}

func (room *Room) createPlayers(pConnections [2]*PlayerConnection) {
	room.players[0] = room.createPlayer(pConnections[0], 0)
	room.players[1] = room.createPlayer(pConnections[1], 1)
}

func (room *Room) createPlayer(pConn *PlayerConnection, playerId int) Player {
	player := CreatePlayer(pConn.id, playerId, room.roomChan)
	pConn.playerChan = player.GetPlayerChan()
	player.StartLoop()

	return player
}

func (room *Room) createGame() {
	room.game = game.CreateGame()

	room.sendMatchStartedMessage(room.players[0])
	room.sendMatchStartedMessage(room.players[1])
}

func (room *Room) sendMatchStartedMessage(player Player) {
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

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: player.connectionID,
		Msg: *matchStartMsg,
	}))
}

func (room *Room) loop() {
	LOOP:
	for {
		select {
		case e := <-room.roomChan:
			eType := e.GetType()

			log.Printf("event in room: Type: %v, ", eType)

			switch eType.GetEventType() {
			case event.EventTypeMove:
				eMove, _ := eType.(EventMove)

				err := room.handleMove(eMove)
				if err != nil {
					log.Println(err)
					log.Printf("cannot handle move for %+v\n", eMove.Player)

					// TODO: refactor handle move function, do i need to end game?
				}

			case event.EventTypeExit:
				eExit, _ := eType.(EventExit)
				opponent := room.GetOpponent(eExit.Player.playerID)

				eExit.Player.EndLoop()
				opponent.EndLoop()

				eExit.RoomUUID = room.GetUUID()
				eExit.OpponentConnId = opponent.connectionID
				room.gameEndWinHandler(eExit.OpponentConnId, eExit.ConnectionId)
				e := event.CreateEvent(eExit)

				room.sendEventToServerChan(e)

			default:
				room.sendEventToServerChan(e)
			}
		case <- room.stopLoop:
			break LOOP
		}
	}
}

func (room *Room) sendEventToServerChan(e event.Event) error {
	if room.serverChan == nil {
		// TODO: assert
		log.Fatalf("I shouldn't be here.")
		panic("I shouldn't be here.")
	}

	room.serverChan <- e
	return nil
}

func (room *Room) handleMove(eMove EventMove) error {
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

	response := new(message.MoveRes) 
	if err != nil {	
		response.Approved = false
		response.Reason = err.Error()
	} else {
		response.Approved = true
	}

	resMsg, err := message.MakeMessage(int(message.TMoveAns), response) 
	if err != nil {
		return err
	}

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: eMove.Player.connectionID,
		Msg: *resMsg,
	}))

	if err != nil {
		return err
	}

	if response.Approved {
		opponent := room.GetOpponent(eMove.Player.playerID)
	
		msgForOpponent, err := message.MakeMessage(message.TOpponentMove, &message.MoveMessage{
			X: eMove.X,
			Y: eMove.Y,
		})
		if err != nil {
			return err
		}

		err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
			ConnectionId: opponent.connectionID,
			Msg: *msgForOpponent,
		}))

		if err != nil {
			return err
		}

		wState := room.game.GetWinState()
		
		if wState == winState.Values.Win {
			err = room.gameEndWinHandler(eMove.Player.connectionID, opponent.connectionID)

			if err != nil {
				return err
			}
		} else if wState == winState.Values.Draw {
			err = room.gameEndDrawHandler(eMove.Player.connectionID, opponent.connectionID)

			if err != nil {
				return err
			}
		}
	} else {
		return errors.New(response.Reason)
	}

	return nil
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

	return &room.players[opponentId]
}

func (room *Room) gameEndWinHandler(winner, loser uuid.UUID) error {
	winMsg, err := message.MakeMessage(message.TWinEvent, &message.WinMessage{
		Status: "win",
		Cause: "",
	})
	if err != nil {
		return err
	} 

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: winner,
		Msg: *winMsg,
	}))

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

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: loser,
		Msg: *loseMsg,
	}))

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

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: c1,
		Msg: *drawMsg,
	}))

	if err != nil {
		return err
	}

	err = room.sendEventToServerChan(event.CreateEvent(EventSendMessage{
		ConnectionId: c2,
		Msg: *drawMsg,
	}))

	if err != nil {
		return err
	}

	return nil
}
