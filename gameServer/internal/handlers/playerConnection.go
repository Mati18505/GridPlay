package handlers

import (
	"log"

	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/message"

	"github.com/google/uuid"
)

type PlayerConnection struct {
	nextHandler Handler
	serverHandler Handler
	uuid uuid.UUID
	connection *connection.Connection
	stopLoop chan bool
	isLoopRunning bool
}

func CreatePlayerConnection(serverHandler Handler, uuid uuid.UUID, conn *connection.Connection) *PlayerConnection {
	assert.NotNil(serverHandler, "server handler was nil")
	assert.NotNil(conn, "connection was nil")

	return &PlayerConnection{
		nextHandler: nil,
		serverHandler: serverHandler,
		uuid: uuid,
		connection: conn,
		stopLoop: make(chan bool),
		isLoopRunning: false,
	}
}

func (playerConn *PlayerConnection) StartLoop() {
	assert.Assert(!playerConn.isLoopRunning, "loop was already running")

	go playerConn.loop()
	playerConn.isLoopRunning = true
}

func (playerConn *PlayerConnection) EndLoop() {
	assert.Assert(playerConn.isLoopRunning, "loop wasn't running")

	playerConn.stopLoop <- true
	playerConn.isLoopRunning = false
}

func (playerConn *PlayerConnection) GetConnection() *connection.Connection {
	assert.NotNil(playerConn.connection, "connection was nil")

	return playerConn.connection;
}

func (playerConn *PlayerConnection) SetNextHandler(nextHandler Handler) {
	assert.NotNil(nextHandler, "next handler was nil")

	playerConn.nextHandler = nextHandler
}

func (pConn *PlayerConnection) Handle(e event.Event) {
	if pConn.nextHandler != nil {
		pConn.nextHandler.Handle(e)
	} else {
		log.Printf("cannot do this while game is not running")

		message := message.MakeMessage(message.TNotAllowedErr, &message.NotAllowedErrMessage{
			Reason: "cannot do this while game is not running",
		})

		pConn.GetConnection().SendMessage(message)
	}
}

func (pConn *PlayerConnection) loop() {
	conn := pConn.GetConnection()
	remoteIP := conn.GetRemoteIP()

	for {
		select {
		case msg := <- conn.GetMessageFromClient():
			log.Printf("playerConnection: received message from %q: Type: %v, ", remoteIP, message.ClientMsg(msg.Type))

			e, err := EventFromMessage(msg)

			if err != nil {
				log.Printf("Unknown type of message, err: %s", err)
				continue
			} 

			log.Printf("playerConnection: Created event %+v", e)
			pConn.Handle(e)

		case <- conn.GetExitChan():
			log.Printf("disconnected from %q\n", remoteIP)

			e := EventDisconnect{
				ConnectionId: pConn.uuid,
			}

			if pConn.nextHandler != nil {
				pConn.nextHandler.Handle(e)
			} else {
				pConn.sendToServerHandler(e)
			}

		case <- pConn.stopLoop:
			return
		}
	}
}

func (pConn *PlayerConnection) sendToServerHandler(e event.Event) {
	assert.NotNil(pConn.serverHandler, "server handler was nil")

	pConn.serverHandler.Handle(e)
}