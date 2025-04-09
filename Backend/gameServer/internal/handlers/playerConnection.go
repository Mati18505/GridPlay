package handlers

import (
	"log/slog"

	"GridPlay/assert"
	"GridPlay/gameServer/internal/connection"
	"GridPlay/gameServer/internal/event"
	"GridPlay/gameServer/message/clientMsg"
	"GridPlay/gameServer/message/serverMsg"

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
	assert.NotNil(playerConn.connection, "connection was nil")

	go playerConn.loop()
	playerConn.isLoopRunning = true

	playerConn.connection.StartReceiving()
}

func (playerConn *PlayerConnection) EndLoop() {
	assert.Assert(playerConn.isLoopRunning, "loop wasn't running")
	assert.NotNil(playerConn.connection, "connection was nil")

	playerConn.connection.StopReceiving()

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
		slog.Info("cannot do this while game is not running")

		message := serverMsg.MakeMessage(serverMsg.TNotAllowedErr, &serverMsg.NotAllowedErrMessage{
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
			slog.Debug("received message from", "ip", remoteIP, "type", clientMsg.MsgType(msg.Type), "data", msg.Data)

			e, err := EventFromClientMessage(msg)

			if err != nil {
				slog.Warn("unknown type of message", "err", err)
				continue
			} 

			slog.Debug("created event", "type", e.GetType(), "event", e)
			pConn.Handle(e)

		case <- conn.GetExitChan():
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

func (pConn *PlayerConnection) GetUUID() uuid.UUID {
	return pConn.uuid
}