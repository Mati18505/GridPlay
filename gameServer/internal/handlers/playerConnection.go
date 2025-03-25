package handlers

import (
	"errors"
	"log"

	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"

	"github.com/google/uuid"
)

type PlayerConnection struct {
	id uuid.UUID
	connection *connection.Connection
	playerChan chan<- event.Event
	serverChan chan<- event.Event
	stopLoop chan bool
}

func CreatePlayerConnection(uuid uuid.UUID, conn *connection.Connection, serverChan chan<- event.Event) *PlayerConnection {
	return &PlayerConnection{
		id: uuid,
		connection: conn,
		playerChan: nil,
		serverChan: serverChan,
		stopLoop: make(chan bool),
	}
}

func (playerConn *PlayerConnection) StartLoop() {
	go playerConn.loop()
}

func (playerConn *PlayerConnection) EndLoop() {
	playerConn.stopLoop <- true
}

func (playerConn *PlayerConnection) GetConnection() *connection.Connection {
	return playerConn.connection;
}

func (pConn *PlayerConnection) loop() {
	conn := pConn.connection
	remoteIP := conn.GetRemoteIP()

	LOOP:
	for {
		select {
		case msg := <- conn.GetMessageFromClient():

			log.Printf("playerConnection: received message from %q: Type: %v, ", remoteIP, message.ClientMsg(msg.Type))

			eType, err := EventTypeFromMessage(msg)

			if err != nil {
				log.Printf("Unknown type of message")
				continue
			} 
			
			e := event.CreateEvent(eType)

			log.Printf("playerConnection: Created event %+v", e)

			err = pConn.sendEventToPlayerChan(e)
				
			if err != nil {
				message, err := message.MakeMessage(message.TNotAllowedErr, &message.NotAllowedErrMessage{
					Reason: "cannot do this while game is not running",
				})

				if err != nil {
					log.Print("cannot make not allowed err message")
					continue;
				}
	
				conn.SendMessage(message)
				log.Printf("cannot do this while game is not running")
			}

		case <- conn.GetExitChan():
			log.Printf("disconnected from %q\n", remoteIP)
			conn.Close()

			e := event.CreateEvent(EventExit{
				ConnectionId: pConn.id,
			})
			err := pConn.sendEventToPlayerChan(e)

			if err != nil {
				e := event.CreateEvent(EventExit{
					ConnectionId: pConn.id,
				})
				pConn.serverChan <- e
			}

			break LOOP
		case <- pConn.stopLoop:
			break LOOP
		}
	}
}

func (pConn *PlayerConnection) sendEventToPlayerChan(e event.Event) error {
	if pConn.playerChan == nil {
		return errors.New("player connection channel does not exist")
	}

	pConn.playerChan <- e
	return nil
}