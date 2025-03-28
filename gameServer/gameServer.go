package gameServer

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/event"
	"TicTacToe/gameServer/internal/handlers"
	. "TicTacToe/gameServer/internal/handlers"
	"TicTacToe/gameServer/message"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server struct {
	connections map[uuid.UUID]*PlayerConnection
	rooms       map[uuid.UUID]*handlers.Room
	mut sync.Mutex
	matcher chan uuid.UUID
	sync *Synchronizer
}

func InitGameServer() *Server {
	srv := &Server{
		connections: make(map[uuid.UUID]*PlayerConnection),
		rooms: make(map[uuid.UUID]*Room),
		matcher: make(chan uuid.UUID, 2),
	}
	srv.sync = CreateSynchronizer(srv)

	go srv.matchMaker()

	return srv
}

func (srv *Server) AddConnection(conn *connection.Connection) {
	assert.NotNil(conn, "connection was nil")
	assert.NotNil(srv.sync, "gameServer sync was nil")

	id := srv.generateUUID()
	pConn := CreatePlayerConnection(srv.sync, id, conn)

	srv.addPlayerConnection(id, pConn)

	conn.StartReceiving()
	pConn.StartLoop()

	log.Printf("connected to %q, uuid:%q\n", conn.GetRemoteIP(), id.String())
}

func (srv *Server) DeleteConnection(id uuid.UUID) {
	conn, err := srv.getConnection(id)
	assert.NoError(err, "player connection does not exist")

	conn.EndLoop()

	srv.removeConnection(id)
}

func (srv *Server) HandleConnection(w http.ResponseWriter, r *http.Request) error {
    socket, err := upgrader.Upgrade(w, r, nil)
	defer r.Body.Close() 

    if err != nil {
        return err
    }

	conn := connection.CreateConnection(socket)
	srv.AddConnection(conn)

	return nil
}

func (srv *Server) Update() {
	assert.NotNil(srv.sync, "server sync was nil")

	srv.updateAllRooms()
	srv.sync.SyncTransferAll()
}

func (srv *Server) updateAllRooms() {
	for _, room := range srv.rooms {
		room.Update()
	}
}

func (srv *Server) addRoom(room *handlers.Room) {
	assert.NotNil(room, "room was nil")

	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.rooms[room.GetUUID()] = room
}

func (srv *Server) createRoom(pConnections [2]*PlayerConnection) *handlers.Room {
	assert.NotNil(srv.sync, "server synchronizer was nil")

	log.Println("creating room")

	uuid := srv.generateUUID()
	room := CreateRoom(srv.sync, pConnections, uuid)

	assert.NotNil(room, "room was nil")
	return room
}

func (srv *Server) removeRoom(roomUUID uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	delete(srv.rooms, roomUUID)
}

func (srv *Server) sendMessage(connId uuid.UUID, msg *message.Message) {
	assert.NotNil(msg, "message was nil")
	
	pConn, err := srv.getConnection(connId)
	assert.NoError(err, "player connection was nil")

	conn := pConn.GetConnection()
	conn.SendMessage(msg)
}

// Blocking
func (srv *Server) matchMaker() {
	for {
		id1, c1 := srv.getConnectionFromMatcher()
		id2, c2 := srv.getConnectionFromMatcher()

		a1 := c1.GetConnection().SendPing() == nil
		a2 := c2.GetConnection().SendPing() == nil

		if a1 && a2 {
			room := srv.createRoom([2]*PlayerConnection{c1, c2})
			
			srv.addRoom(room)
		} else if !a1 {
			srv.matcher <- id2
		} else if !a2 {
			srv.matcher <- id1
		} else {
			continue
		}
	}
}

func (srv *Server) getConnectionFromMatcher() (uuid.UUID, *PlayerConnection) {
	id := <-srv.matcher
	conn, err := srv.getConnection(id)

	assert.NoError(err, "player connection was nil")
	return id, conn
}

func (srv *Server) eventNotHandled(e event.Event) {
	// TODO: Check if it's logged correctly.
	log.Printf("event not handled: %+v", e)
}

func (srv *Server) Handle(e event.Event) {
	eType := e.GetType()

	switch eType {
	case event.EventTypeSendMessage:
		eSendMessage, ok := e.(EventSendMessage)
		assert.Assert(ok, "type assertion failed for event sendMessage")

		srv.sendMessage(eSendMessage.ConnectionId, &eSendMessage.Msg)
		
	case event.EventTypeExit:
		eExit, ok := e.(EventExit)
		assert.Assert(ok, "type assertion failed for event exit")

		srv.matcher <- eExit.OpponentConnId
		srv.DeleteConnection(eExit.ConnectionId)

		log.Println("removing room")
		srv.removeRoom(eExit.RoomUUID)

	default:
		srv.eventNotHandled(e)
	}
}

var upgrader = websocket.Upgrader {
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (srv *Server) generateUUID() uuid.UUID {
	uuid, err := uuid.NewUUID()

	assert.NoError(err, "uuid generation error")
	return uuid
}

func (srv *Server) addPlayerConnection(uuid uuid.UUID, pConn *PlayerConnection) {
	assert.NotNil(pConn, "player connection was nil")

	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.connections[uuid] = pConn
	srv.matcher <- uuid
}

func (srv *Server) removeConnection(uuid uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	delete(srv.connections, uuid)
}

func (srv *Server) getConnection(id uuid.UUID) (*PlayerConnection, error) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	conn, ok := srv.connections[id]

	if !ok {
		return nil, errors.New("connection does not exist")
	}

	return conn, nil
}