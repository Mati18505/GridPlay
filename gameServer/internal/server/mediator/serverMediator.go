package mediator

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/internal/handlers"
	"TicTacToe/gameServer/internal/server/matchmaker"
	"TicTacToe/gameServer/internal/server/serverData"
	"TicTacToe/gameServer/internal/server/serverEvents"
	"TicTacToe/gameServer/message"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type ServerMediator struct {
	handler *handlers.ServerHandler
	matchmaker *matchmaker.Matchmaker
	serverData *serverData.ServerData
}

func CreateServerMediator() *ServerMediator {
	mediator := &ServerMediator{}

	mediator.handler = handlers.CreateServerHandler(mediator)
	mediator.matchmaker = matchmaker.CreateMatchMaker(mediator)
	mediator.serverData = serverData.CreateServerData()

	return mediator
}

func (mediator *ServerMediator) StartLoop() {
	assert.NotNil(mediator.matchmaker, "matchmaker was nil")

	mediator.matchmaker.StartLoop()
}

func (mediator *ServerMediator) StopLoop() {
	assert.NotNil(mediator.matchmaker, "matchmaker was nil")

	mediator.matchmaker.EndLoop()
}

func (mediator *ServerMediator) Notify(e serverEvents.MediatorEvent) {
	handled := false

	if e.Sender == serverEvents.ServerHandler {
		handled = mediator.FromServerHandler(e.Event)
	} else if e.Sender == serverEvents.Matchmaker {
		handled = mediator.FromMatchmaker(e.Event)
	} 

	if !handled {
		mediator.eventNotHandled(e)
	}
}

func (mediator *ServerMediator) FromServerHandler(e event.Event) bool {
	eType := e.GetType()
	switch eType {
	case event.EventTypeSendMessage:
		eSendMessage, ok := e.(handlers.EventSendMessage)
		assert.Assert(ok, "type assertion failed for event sendMessage")

		err := mediator.SendMessage(eSendMessage.ConnectionId, eSendMessage.Msg)

		if err != nil {
			fmt.Printf("Cannot send message to connection %v, %s", eSendMessage.ConnectionId, err)
		}
		
	case event.EventTypeDisconnect:
		eExit, ok := e.(handlers.EventDisconnect)
		assert.Assert(ok, "type assertion failed for event exit")

		mediator.DeleteConnection(eExit.ConnectionId)

	case event.EventTypeRemoveRoom:
		eRemoveRoom, ok := e.(handlers.EventRemoveRoom)
		assert.Assert(ok, "type assertion failed for event remove room")
		assert.NotNil(mediator.matchmaker, "matchmaker was nil")
		assert.NotNil(mediator.serverData, "serverData was nil")

		mediator.DeleteConnection(eRemoveRoom.ConnectionId)
		mediator.matchmaker.Add(eRemoveRoom.OpponentConnId)

		log.Println("removing room")
		mediator.serverData.RemoveRoom(eRemoveRoom.RoomUUID)
	default:
		return false
	}

	return true
}

func (mediator *ServerMediator) FromMatchmaker(e event.Event) bool {
	eType := e.GetType()
	switch eType {
	case event.EventTypePlayersMatched:
		ePlayersMatched, ok := e.(matchmaker.EventPlayersMatched)
		assert.Assert(ok, "type assertion failed for event players matched")
		assert.NotNil(mediator.matchmaker, "matchmaker was nil")
		assert.NotNil(mediator.serverData, "serverData was nil")

		ids := ePlayersMatched.Ids
		conns := [2]*handlers.PlayerConnection{}
		var confirm[2]bool

		for i := range conns {
			conn, err := mediator.serverData.GetConnection(ids[i])
			
			// TODO: funcition connection confirm
			if err != nil || conn.GetConnection().SendPing() != nil {
				confirm[i] = false
				continue
			} 

			conns[i] = conn
			confirm[i] = true
		}

		if confirm[0] && confirm[1] {
			room := mediator.CreateRoom(conns)

			mediator.serverData.AddRoom(room)
		} else if confirm[0] {
			mediator.matchmaker.Add(ids[0])
		} else if confirm[1] {
			mediator.matchmaker.Add(ids[1])
		}
	default:
		return false
	}

	return true
}

func (mediator *ServerMediator) CreateRoom(pConnections [2]*handlers.PlayerConnection) *handlers.Room {
	assert.NotNil(mediator.handler, "server handler was nil")

	log.Println("creating room")

	uuid := mediator.GenerateUUID()
	room := handlers.CreateRoom(mediator.handler.GetSync(), pConnections, uuid)

	assert.NotNil(room, "room was nil")
	return room
}

func (mediator *ServerMediator) SendMessage(connId uuid.UUID, msg message.Message) error {
	assert.NotNil(mediator.serverData, "server data was nil")
	assert.NotNil(msg, "message was nil")
	
	pConn, err := mediator.serverData.GetConnection(connId)

	if err != nil {
		return err
	}

	conn := pConn.GetConnection()
	conn.SendMessage(msg)

	return nil
}

func (mediator *ServerMediator) GenerateUUID() uuid.UUID {
	uuid, err := uuid.NewUUID()

	assert.NoError(err, "uuid generation error")
	return uuid
}

func (mediator *ServerMediator) AddConnection(conn *connection.Connection) {
	assert.NotNil(mediator.serverData, "server data was nil")
	assert.NotNil(mediator.handler, "server handler was nil")
	assert.NotNil(mediator.matchmaker, "server handler was nil")
	assert.NotNil(conn, "connection was nil")

	id := mediator.GenerateUUID()
	pConn := handlers.CreatePlayerConnection(mediator.handler.GetSync(), id, conn)

	mediator.serverData.AddPlayerConnection(id, pConn)
	mediator.matchmaker.Add(id)

	conn.StartReceiving()
	pConn.StartLoop()

	log.Printf("connected to %q, uuid:%q\n", conn.GetRemoteIP(), id.String())
}

func (mediator *ServerMediator) DeleteConnection(id uuid.UUID) {
	assert.NotNil(mediator.serverData, "server data was nil")

	conn, err := mediator.serverData.GetConnection(id)
	assert.NoError(err, "player connection does not exist")

	conn.EndLoop()

	mediator.serverData.RemoveConnection(id)
}

func (mediator *ServerMediator) Update() {
	assert.NotNil(mediator.handler, "server handler was nil")

	mediator.updateAllRooms()
	mediator.handler.GetSync().SyncTransferAll()
}

func (mediator *ServerMediator) updateAllRooms() {
	assert.NotNil(mediator.serverData, "server data was nil")

	mediator.serverData.ForEachRoom(func(room *handlers.Room) {
		room.Update()
	})
}

func (mediator *ServerMediator) eventNotHandled(e serverEvents.MediatorEvent) {
	// TODO: Check if it's logged correctly.
	log.Printf("event not handled: %+v", e)
}