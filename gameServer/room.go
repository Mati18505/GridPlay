package gameServer

import (
	"TicTacToe/game"
	"TicTacToe/game/winState"
	"errors"
	"log"

	"github.com/google/uuid"
)

type room struct {
	uuid uuid.UUID
	game        *game.Game
	players [2]player
	serverChan chan<- Event
	roomChan chan Event
	stopLoop chan bool
}

func CreateRoom(pConnections [2]*playerConnection, uuid uuid.UUID, serverChan chan<- Event) *room {
	room := &room{
		uuid: uuid,
		serverChan: serverChan,
		roomChan: make(chan Event),
		stopLoop: make(chan bool),
	}

	room.createPlayers(pConnections)
	room.createGame()

	return room
}

func (room *room) StartLoop() {
	go room.loop()
}

func (room *room) EndLoop() {
	room.stopLoop <- true
}

func (room *room) GetUUID() uuid.UUID {
	return room.uuid
}

func (room *room) createPlayers(pConnections [2]*playerConnection) {
	room.players[0] = room.createPlayer(pConnections[0], 0)
	room.players[1] = room.createPlayer(pConnections[1], 1)
}

func (room *room) createPlayer(pConn *playerConnection, playerId int) player {
	player := CreatePlayer(pConn.id, playerId, room.roomChan)
	pConn.playerChan = player.GetPlayerChan()
	player.StartLoop()

	return player
}

func (room *room) createGame() {
	room.game = game.CreateGame()

	room.sendMatchStartedMessage(room.players[0])
	room.sendMatchStartedMessage(room.players[1])
}

func (room *room) sendMatchStartedMessage(player player) {
	gamePlayer := room.game.GetPlayerWithId(player.playerID)
	playerChar := gamePlayer.GetChar()
	opponentChar := game.OpponentChar(playerChar)

	matchStartMsg, err := MakeMessage(MatchStarted, &matchStarted{
		Char: rune(playerChar),
		OpponentChar: rune(opponentChar),
	})

	if err != nil {
		log.Print("cannot make match started message")
		// TODO: what now?
	}

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: player.connectionID,
		msg: *matchStartMsg,
	}))
}

func (room *room) loop() {
	LOOP:
	for {
		select {
		case e := <-room.roomChan:
			eType := e.GetType()

			log.Printf("event in room: Type: %v, ", eType)

			switch eType.GetEventType() {
			case EventTypeMove:
				eMove, _ := eType.(EventMove)

				err := room.handleMove(eMove)
				if err != nil {
					log.Println(err)
					log.Printf("cannot handle move for %+v\n", eMove.player)

					// TODO: refactor handle move function, do i need to end game?
				}

			case EventTypeExit:
				eExit, _ := eType.(EventExit)
				opponent := room.GetOpponent(eExit.player.playerID)

				eExit.player.EndLoop()
				opponent.EndLoop()

				eExit.roomUUID = room.GetUUID()
				eExit.opponentConnId = opponent.connectionID
				room.gameEndWinHandler(eExit.opponentConnId, eExit.connectionId)
				e := CreateEvent(eExit)

				room.sendEventToServerChan(e)

			default:
				room.sendEventToServerChan(e)
			}
		case <- room.stopLoop:
			break LOOP
		}
	}
}

func (room *room) sendEventToServerChan(e Event) error {
	if room.serverChan == nil {
		// TODO: assert
		log.Fatalf("I shouldn't be here.")
		panic("I shouldn't be here.")
	}

	room.serverChan <- e
	return nil
}

func (room *room) handleMove(eMove EventMove) error {
	// TODO: assert if game don't exist ? (maybe always exist)
	var err error
	currPlayer := room.game.GetCurrentRoundPlayer()

	gamePlayer := room.game.GetPlayerWithId(eMove.player.playerID)

	if err != nil {
		return err
	}

	if currPlayer == gamePlayer {
		err = room.game.Move(game.Pos{X: eMove.x, Y: eMove.y})
	} else {
		err = errors.New("not your round, dummy")
	}

	response := new(moveRes) 
	if err != nil {	
		response.Approved = false
		response.Reason = err.Error()
	} else {
		response.Approved = true
	}

	resMsg, err := MakeMessage(int(MoveAns), response) 
	if err != nil {
		return err
	}

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: eMove.player.connectionID,
		msg: *resMsg,
	}))

	if err != nil {
		return err
	}

	if response.Approved {
		opponent := room.GetOpponent(eMove.player.playerID)
	
		msgForOpponent, err := MakeMessage(OpponentMove, &moveMessage{
			X: eMove.x,
			Y: eMove.y,
		})
		if err != nil {
			return err
		}

		err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
			connectionId: opponent.connectionID,
			msg: *msgForOpponent,
		}))

		if err != nil {
			return err
		}

		wState := room.game.GetWinState()
		
		if wState == winState.Values.Win {
			err = room.gameEndWinHandler(eMove.player.connectionID, opponent.connectionID)

			if err != nil {
				return err
			}
		} else if wState == winState.Values.Draw {
			err = room.gameEndDrawHandler(eMove.player.connectionID, opponent.connectionID)

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
func (room *room) GetOpponent(playerID int) *player {
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

func (room *room) gameEndWinHandler(winner, loser uuid.UUID) error {
	winMsg, err := MakeMessage(WinEvent, &winMessage{
		Status: "win",
		Cause: "",
	})
	if err != nil {
		return err
	} 

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: winner,
		msg: *winMsg,
	}))

	if err != nil {
		return err
	}
	
	loseMsg, err := MakeMessage(WinEvent, &winMessage{
		Status: "lose",
		Cause: "",
	})

	if err != nil {
		return err
	} 

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: loser,
		msg: *loseMsg,
	}))

	if err != nil {
		return err
	}

	return nil
}

func (room *room) gameEndDrawHandler(c1, c2 uuid.UUID) error {
	drawMsg, err := MakeMessage(WinEvent, &winMessage{
		Status: "draw",
		Cause: "",
	})
	if err != nil {
		return err
	} 

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: c1,
		msg: *drawMsg,
	}))

	if err != nil {
		return err
	}

	err = room.sendEventToServerChan(CreateEvent(EventSendMessage{
		connectionId: c2,
		msg: *drawMsg,
	}))

	if err != nil {
		return err
	}

	return nil
}
