package gameServer

import (
	"TicTacToe/game"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

type room struct {
	game        *game.Game
}

type Server struct {
	connections map[*websocket.Conn]*Connection
	rooms       []*room
	mut sync.Mutex
	matcher chan *Connection
}

func InitGameServer() *Server {
	srv := &Server{
		connections: make(map[*websocket.Conn]*Connection),
		rooms: make([]*room, 0),
		matcher: make(chan *Connection, 2),
	}
	go srv.matchMaker()
	return srv
}

func (srv *Server) AddConnection(id *websocket.Conn, conn *Connection) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	_, exist := srv.connections[id]
	if exist {
		return errors.New("Connection with this id is already in the map")
	}

	srv.connections[id] = conn
	srv.matcher <- conn

	log.Printf("connected to %q\n", conn.GetRemoteIP())

	go conn.receiveMessages()
	go srv.routeMessages(conn)

	return nil
}

func (srv *Server) DeleteConnection(id *websocket.Conn) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	delete(srv.connections, id)
}

func (srv *Server) addGame(game *game.Game) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	log.Println("creating room")
	room := &room{
		game: game,
	}
	srv.rooms = append(srv.rooms, room)
}

// Blocking
func (srv *Server) matchMaker() {
	for {
		p1 := <-srv.matcher
		p2 := <-srv.matcher

		matchStartMsg := &message{
			Type: MatchStarted,
		}

		a1 := p1.sendPing() == nil
		a2 := p2.sendPing() == nil

		if a1 && a2 {
			game := game.CreateGame(p1.player, p2.player)
			srv.addGame(game)
			p1.sendMessage(matchStartMsg)
			p2.sendMessage(matchStartMsg)
			game.EndGame = srv.gameEndHandler(p1, p2)
		} else if !a1 {
			srv.matcher <- p2
		} else if !a2 {
			srv.matcher <- p1
		} else {
			continue
		}
	}
}

// Blocking
func (srv *Server) routeMessages(conn *Connection) {
	BREAK:
	for {
		select {
			case msg := <- conn.messageFromClient:

				log.Printf("message from %q: Type: %v, ", conn.GetRemoteIP(), ClientMsg(msg.Type))

				switch ClientMsg(msg.Type) {
				case Move:
					data, err := ParseMessage[moveMessage](msg)
					if err != nil {
						log.Println(err)
						continue
					}

					err = srv.handleMove(conn, data)
					if err != nil {
						log.Println(err)
						log.Printf("cannot handle move for %q disconnecting...\n", conn.GetRemoteIP())
						break BREAK
					}

				default:
					log.Printf("received unknown type (%v) of message from %q ignoring...\n", msg.Type, conn.GetRemoteIP())
					continue
				}
			case <- conn.exitChan:
				log.Printf("disconnected from %q\n", conn.GetRemoteIP())
				srv.DeleteConnection(conn.socket)
				conn.close()
				// Send win message to opponent
				// Send opponent to lobby
				//delete rooms[roomIdx]
				break BREAK
		}
	}
}

func (srv *Server) handleMove(conn *Connection, msg *moveMessage) error {
	log.Printf("Data: %q\n", msg)
	pos := game.Pos{ X: msg.X, Y: msg.Y }

	err := conn.player.Move(pos)
	response := new(moveRes) 
	if err != nil {	
		response.Approved = false
		response.Reason = err.Error()
	} else {
		response.Approved = true
	}
	resMsg, err := MakeMessage(int(MoveAns), response) 
	if err != nil {
		return err
	}

	err = conn.sendMessage(resMsg)
	return err
}

func (srv *Server) gameEndHandler(p1, p2 *Connection) func (winner int) {
	return func (winnerID int)  {
		var winner *Connection
		var looser *Connection

		if p1.player.GetID() == winnerID {
			winner = p1
			looser = p2
		} else {
			winner = p2
			looser = p1
		}
		
		winMsg, err := MakeMessage(WinEvent, &winMessage{
			Status: "win",
			Cause: "",
		})
		if err != nil {
			fmt.Println(err)
		} else {
			winner.sendMessage(winMsg)
		}
		
		looseMsg, err := MakeMessage(WinEvent, &winMessage{
			Status: "loose",
			Cause: "",
		})
		if err != nil {
			fmt.Println(err)
		} else {
			looser.sendMessage(looseMsg) 
		}
	}
}

func (srv *Server) GetRoomWithGame(g *game.Game) {
	slices.IndexFunc(srv.rooms, func (room *room) bool {
		return room.game == g;		
	})
}