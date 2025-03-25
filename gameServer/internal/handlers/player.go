package handlers

import (
	"log"

	"github.com/google/uuid"

	"TicTacToe/gameServer/internal/event"
)

type Player struct {
	connectionID uuid.UUID
	playerID int
	roomChan chan<- event.Event
	playerChan chan event.Event
	stopLoop chan bool
}

func CreatePlayer(connId uuid.UUID, playerId int, roomChan chan<- event.Event) Player {
	return Player{
		connectionID: connId,
		playerID: playerId,
		roomChan: roomChan,
		playerChan: make(chan event.Event),
		stopLoop: make(chan bool),
	}
}

func (player *Player) StartLoop() {
	go player.loop()
}

func (player *Player) EndLoop() {
	player.stopLoop <- true
}

func (player *Player) GetPlayerChan() chan<-event.Event {
	return player.playerChan
}

func (player *Player) loop() {
	LOOP:
	for {
		select {
		case e := <-player.playerChan:
			eType := e.GetType()

			log.Printf("event in player: Type: %v, ", eType)

			switch eType.GetEventType() {
			case event.EventTypeMove:
				eMove, _ := eType.(EventMove)

				eMove.Player = player
				e := event.CreateEvent(eMove)

				player.sendEventToRoomChan(e)

			case event.EventTypeExit:
				eExit, _ := eType.(EventExit)

				eExit.Player = player
				e := event.CreateEvent(eExit)

				player.sendEventToRoomChan(e)

			default:
				player.sendEventToRoomChan(e)
			}
		case <- player.stopLoop:
			break LOOP
		}
	}
}

func (player *Player) sendEventToRoomChan(e event.Event) {
	if player.roomChan == nil {
		// TODO: assert
		log.Fatalf("I shouldn't be here.")
		panic("I shouldn't be here.")
	}

	player.roomChan <- e
}