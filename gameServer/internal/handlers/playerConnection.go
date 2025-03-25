package handlers

import (
	"log"

	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"

	"github.com/google/uuid"
)

type PlayerConnection struct {
	nextHandler Handler
	serverHandler Handler
	id uuid.UUID
	connection *connection.Connection
	stopLoop chan bool
}

func CreatePlayerConnection(serverHandler Handler, uuid uuid.UUID, conn *connection.Connection) *PlayerConnection {
	return &PlayerConnection{
		nextHandler: nil,
		serverHandler: serverHandler,
		id: uuid,
		connection: conn,
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

func (playerConn *PlayerConnection) SetNextHandler(nextHandler Handler) {
	playerConn.nextHandler = nextHandler
}

func (pConn *PlayerConnection) Handle(e event.Event) {
	if pConn.nextHandler != nil {
		pConn.nextHandler.Handle(e)
	} else {
		log.Printf("cannot do this while game is not running")

		message, err := message.MakeMessage(message.TNotAllowedErr, &message.NotAllowedErrMessage{
			Reason: "cannot do this while game is not running",
		})

		if err != nil {
			log.Print("cannot make not allowed err message")
			// TODO: What now?
		}

		pConn.connection.SendMessage(message)
	}
}

func (pConn *PlayerConnection) loop() {
	conn := pConn.connection
	remoteIP := conn.GetRemoteIP()

	LOOP:
	for {
		select {
		case msg := <- conn.GetMessageFromClient():

			log.Printf("playerConnection: received message from %q: Type: %v, ", remoteIP, message.ClientMsg(msg.Type))

			e, err := EventFromMessage(msg)

			if err != nil {
				log.Printf("Unknown type of message")
				continue
			} 

			log.Printf("playerConnection: Created event %+v", e)
			pConn.Handle(e)

		case <- conn.GetExitChan():
			log.Printf("disconnected from %q\n", remoteIP)
			conn.Close()

			e := EventExit{
				ConnectionId: pConn.id,
			}

			if pConn.nextHandler != nil {
				pConn.nextHandler.Handle(e)
			} else {
				pConn.serverHandler.Handle(e)
			}

			break LOOP

		case <- pConn.stopLoop:
			break LOOP
		}
	}
}