package gameServer

import (
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
	matcher chan *PlayerConnection
}

func InitGameServer() *Server {
	srv := &Server{
		connections: make(map[uuid.UUID]*PlayerConnection),
		rooms: make(map[uuid.UUID]*Room),
		matcher: make(chan *PlayerConnection, 2),
	}
	go srv.matchMaker()

	return srv
}

func (srv *Server) AddConnection(conn *connection.Connection) error {
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

	pConn := CreatePlayerConnection(srv, id, conn)

	srv.connections[id] = pConn
	srv.matcher <- pConn

	log.Printf("connected to %q, uuid:%q\n", conn.GetRemoteIP(), id.String())

	go conn.ReceiveMessages()
	pConn.StartLoop()

	return nil
}

func (srv *Server) DeleteConnection(id uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.connections[id].EndLoop()

	delete(srv.connections, id)
}

func (srv *Server) HandleConnection(w http.ResponseWriter, r *http.Request) error {
    socket, err := upgrader.Upgrade(w, r, nil)
	
    if err != nil {
		r.Body.Close() // Is it needed?
        return err
    }

	conn := connection.CreateConnection(socket)
	err = srv.AddConnection(conn)

	return err
}

func (srv *Server) addRoom(room *handlers.Room) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	srv.rooms[room.GetUUID()] = room
}

func (srv *Server) createRoom(pConnections [2]*PlayerConnection) (*handlers.Room, error) {
	log.Println("creating room")

	uuid, err := uuid.NewUUID()
	if err != nil {
		return nil, errors.New("cannot generate uuid for this room")
	}

	room := CreateRoom(srv, pConnections, uuid)

	return room, nil
}

func (srv *Server) removeRoom(roomUUID uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	delete(srv.rooms, roomUUID)
}

func (srv *Server) sendMessage(connId uuid.UUID, msg *message.Message) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	conn := srv.connections[connId].GetConnection()
	// TODO: Check this error!
	conn.SendMessage(msg)
}

// Blocking
func (srv *Server) matchMaker() {
	for {
		c1 := <-srv.matcher
		c2 := <-srv.matcher

		a1 := c1.GetConnection().SendPing() == nil
		a2 := c2.GetConnection().SendPing() == nil

		if a1 && a2 {
			room, err := srv.createRoom([2]*PlayerConnection{c1, c2})
			
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

func (srv *Server) eventNotHandled(e event.Event) {
	// TODO: Check if it's logged correctly.
	log.Printf("event not handled: %+v", e)
}

func (srv *Server) Handle(e event.Event) {
	eType := e.GetType()

	switch eType {
	case event.EventTypeSendMessage:
		eSendMessage, _ := e.(EventSendMessage)

		srv.sendMessage(eSendMessage.ConnectionId, &eSendMessage.Msg)
		
	case event.EventTypeExit:
		eExit, _ := e.(EventExit)

		// TODO: Get connection via mut lock.
		srv.matcher <- srv.connections[eExit.OpponentConnId]
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