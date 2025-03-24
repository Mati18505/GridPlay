package gameServer

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Server struct {
	connections map[uuid.UUID]*playerConnection
	rooms       map[uuid.UUID]*room
	mut sync.Mutex
	matcher chan *playerConnection
	serverChan chan Event
	stopLoop chan bool
}

func InitGameServer() *Server {
	srv := &Server{
		connections: make(map[uuid.UUID]*playerConnection),
		rooms: make(map[uuid.UUID]*room),
		matcher: make(chan *playerConnection, 2),
		serverChan: make(chan Event),
		stopLoop: make(chan bool),
	}
	return srv
}

func (srv *Server) StartLoop() {
	go srv.matchMaker()
	go srv.loop()
}

func (srv *Server) EndLoop() {
	srv.stopLoop <- true
}

func (srv *Server) AddConnection(conn *Connection) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return errors.New("cannot generate uuid for this connection")
	}

	_, exist := srv.connections[id]
	if exist {
		return errors.New("Connection with this id is already in the map")
	}

	pConn := CreatePlayerConnection(id, conn, srv.serverChan)

	srv.connections[id] = pConn
	srv.matcher <- pConn

	log.Printf("connected to %q, uuid:%q\n", conn.GetRemoteIP(), id.String())

	go conn.receiveMessages()
	pConn.StartLoop()

	return nil
}

func (srv *Server) DeleteConnection(id uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.connections[id].EndLoop()

	delete(srv.connections, id)
}

func (srv *Server) addRoom(room *room) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.rooms[room.GetUUID()] = room
}

func (srv *Server) createRoom(pConnections [2]*playerConnection) (*room, error) {
	log.Println("creating room")

	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.New("cannot generate uuid for this room")
	}

	room := CreateRoom(pConnections, uuid, srv.serverChan)

	room.StartLoop()
	return room, nil
}

func (srv *Server) removeRoom(roomUUID uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.rooms[roomUUID].EndLoop()

	delete(srv.rooms, roomUUID)
}

func (srv *Server) sendMessage(connId uuid.UUID, msg *message) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	conn := srv.connections[connId].connection
	conn.sendMessage(msg)
}

// Blocking
func (srv *Server) matchMaker() {
	for {
		c1 := <-srv.matcher
		c2 := <-srv.matcher

		a1 := c1.connection.sendPing() == nil
		a2 := c2.connection.sendPing() == nil

		if a1 && a2 {
			room, err := srv.createRoom([2]*playerConnection{c1, c2})
			
			if err != nil {
				log.Printf("cannot create room, err: %s", err)
				// TODO: what now?
			}
			srv.addRoom(room)
		} else if !a1 {
			srv.matcher <- c2
		} else if !a2 {
			srv.matcher <- c1
		} else {
			continue
		}
	}
}

func (srv *Server) eventNotHandled(e Event) {
	// TODO: Check if it's logged correctly.
	log.Printf("event not handled: %+v", e)
}

func (srv *Server) loop() {
	LOOP:
	for {
		select {
		case e := <-srv.serverChan:
			eType := e.GetType()

			switch eType.GetEventType() {
			case EventTypeSendMessage:
				eSendMessage, _ := eType.(EventSendMessage)

				srv.sendMessage(eSendMessage.connectionId, &eSendMessage.msg)
				
			case EventTypeExit:
				eExit, _ := eType.(EventExit)

				srv.matcher <- srv.connections[eExit.opponentConnId]
				srv.DeleteConnection(eExit.connectionId)
				log.Println("removing room")
				srv.removeRoom(eExit.roomUUID)

			default:
				srv.eventNotHandled(e)
			}
		case <- srv.stopLoop:
			break LOOP
		}
	}
}