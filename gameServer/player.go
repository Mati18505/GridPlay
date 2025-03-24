package gameServer

import (
	"log"

	"github.com/google/uuid"
)

type player struct {
	connectionID uuid.UUID
	playerID int
	roomChan chan<- Event
	playerChan chan Event
	stopLoop chan bool
}

func CreatePlayer(connId uuid.UUID, playerId int, roomChan chan<- Event) player {
	return player{
		connectionID: connId,
		playerID: playerId,
		roomChan: roomChan,
		playerChan: make(chan Event),
		stopLoop: make(chan bool),
	}
}

func (player *player) StartLoop() {
	go player.loop()
}

func (player *player) EndLoop() {
	player.stopLoop <- true
}

func (player *player) GetPlayerChan() chan<-Event {
	return player.playerChan
}

func (player *player) loop() {
	LOOP:
	for {
		select {
		case e := <-player.playerChan:
			eType := e.GetType()

			log.Printf("event in player: Type: %v, ", eType)

			switch eType.GetEventType() {
			case EventTypeMove:
				eMove, _ := eType.(EventMove)

				eMove.player = player
				e := CreateEvent(eMove)

				player.sendEventToRoomChan(e)

			case EventTypeExit:
				eExit, _ := eType.(EventExit)

				eExit.player = player
				e := CreateEvent(eExit)

				player.sendEventToRoomChan(e)

			default:
				player.sendEventToRoomChan(e)
			}
		case <- player.stopLoop:
			break LOOP
		}
	}
}

func (player *player) sendEventToRoomChan(e Event) {
	if player.roomChan == nil {
		// TODO: assert
		log.Fatalf("I shouldn't be here.")
		panic("I shouldn't be here.")
	}

	player.roomChan <- e
}