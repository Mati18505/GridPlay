package mediator

import (
	"GridPlay/assert"
	"GridPlay/gameServer/internal/connection"
	"GridPlay/gameServer/internal/event"
	"GridPlay/gameServer/internal/handlers"
	"GridPlay/gameServer/internal/handlers/room"
	"GridPlay/gameServer/internal/server/matchmaker"
	"GridPlay/gameServer/internal/server/serverData"
	"GridPlay/gameServer/internal/server/serverEvents"
	"GridPlay/gameServer/message"
	"GridPlay/gameServer/message/serverMsg"
	"log/slog"

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
	slog.Debug("event in server mediator", "Type", eType, "event", e)

	switch eType {
	case event.EventTypeSendMessage:
		eSendMessage, ok := e.(handlers.EventSendMessage)
		assert.Assert(ok, "type assertion failed for event sendMessage")

		err := mediator.SendMessage(eSendMessage.ConnectionId, eSendMessage.Msg)

		if err != nil {
			slog.Warn("Cannot send message to connection", "ip", eSendMessage.ConnectionId, "err", err)
		}
		
	case event.EventTypeDisconnect:
		eDisconnect, ok := e.(handlers.EventDisconnect)
		assert.Assert(ok, "type assertion failed for event disconnect")

		mediator.DeleteConnection(eDisconnect.ConnectionId)

	case event.EventTypeRemoveRoom:
		eRemoveRoom, ok := e.(handlers.EventRemoveRoom)
		assert.Assert(ok, "type assertion failed for event remove room")

		mediator.RemoveRoom(eRemoveRoom.RoomUUID)
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

		slog.Debug("players matched event", "id1", ePlayersMatched.Ids[0], "id2", ePlayersMatched.Ids[1])

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

func (mediator *ServerMediator) CreateRoom(pConnections [2]*handlers.PlayerConnection) *room.Room {
	assert.NotNil(mediator.handler, "server handler was nil")

	uuid := mediator.GenerateUUID()
	room := room.CreateRoom(mediator.handler.GetSync(), pConnections, uuid)

	slog.Info("created room", "uuid", uuid.String())

	assert.NotNil(room, "room was nil")
	return room
}

func (mediator *ServerMediator) RemoveRoom(uuid uuid.UUID) {
	assert.NotNil(mediator.serverData, "serverData was nil")

	_, err := mediator.serverData.GetRoom(uuid)
	assert.NoError(err, "room does not exist")

	slog.Info("removing room", "uuid", uuid)
	mediator.serverData.RemoveRoom(uuid)
}

func (mediator *ServerMediator) SendMessage(connId uuid.UUID, msg message.Message) error {
	assert.NotNil(mediator.serverData, "server data was nil")
	assert.NotNil(msg, "message was nil")

	pConn, err := mediator.serverData.GetConnection(connId)
	ip := "unaccessible"

	if pConn != nil {
		ip = pConn.GetConnection().GetRemoteIP()
	}

	slog.Debug("sending message", "ip", ip, "type", serverMsg.MsgType(msg.Type), "data", msg.Data)

	if err != nil {
		return err
	}

	conn := pConn.GetConnection()
	conn.SendMessage(msg)

	err = conn.GetLastError()
	return err
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

	pConn.StartLoop()
	mediator.matchmaker.Add(id)

	slog.Info("connected to", "ip", conn.GetRemoteIP(), "uuid", id.String())
}

func (mediator *ServerMediator) DeleteConnection(id uuid.UUID) {
	assert.NotNil(mediator.serverData, "server data was nil")

	conn, err := mediator.serverData.GetConnection(id)
	assert.NoError(err, "player connection does not exist")

	slog.Debug("removing connection", "ip", conn.GetConnection().GetRemoteIP())

	conn.EndLoop()

	mediator.serverData.RemoveConnection(id)
}

func (mediator *ServerMediator) AddConnectionToMatchmaker(uuid uuid.UUID) {
	_, err := mediator.serverData.GetConnection(uuid)
	assert.NoError(err, "connection does not exist")

	slog.Debug("adding player to matchmaker", "uuid", uuid.String())
	mediator.matchmaker.Add(uuid)
}

func (mediator *ServerMediator) Update() {
	assert.NotNil(mediator.handler, "server handler was nil")

	mediator.updateAllRooms()
	mediator.handler.GetSync().SyncTransferAll()
}

func (mediator *ServerMediator) updateAllRooms() {
	assert.NotNil(mediator.serverData, "server data was nil")

	mediator.serverData.ForEachRoom(func(room *room.Room) {
		room.Update()
	})
}

func (mediator *ServerMediator) eventNotHandled(e serverEvents.MediatorEvent) {
	slog.Error("server event not handled", "Sender", e.Sender, "Type", e.Event.GetType())
}