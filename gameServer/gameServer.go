package gameServer

import (
	"TicTacToe/game"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/google/uuid"
)

type room struct {
	game        *game.Game
	connections [2]uuid.UUID
}

type Server struct {
	connections map[uuid.UUID]*Connection
	rooms       []*room
	mut sync.Mutex
	matcher chan *Connection
}

func InitGameServer() *Server {
	srv := &Server{
		connections: make(map[uuid.UUID]*Connection),
		rooms: make([]*room, 0),
		matcher: make(chan *Connection, 2),
	}
	go srv.matchMaker()
	return srv
}

func (srv *Server) AddConnection(conn *Connection) error {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return errors.New("cannot generate uuid for this connection")
	}

	conn.id = id

	_, exist := srv.connections[id]
	if exist {
		return errors.New("Connection with this id is already in the map")
	}

	srv.connections[id] = conn
	srv.matcher <- conn

	log.Printf("connected to %q, uuid:%q\n", conn.GetRemoteIP(), id.String())

	go conn.receiveMessages()
	go srv.routeMessages(conn)

	return nil
}

func (srv *Server) DeleteConnection(id uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	delete(srv.connections, id)
}

func (srv *Server) addGame(game *game.Game, connections [2]uuid.UUID) {
	srv.mut.Lock()
	defer srv.mut.Unlock()

	log.Println("creating room")
	room := &room{
		game: game,
		connections: connections,
	}
	srv.rooms = append(srv.rooms, room)
}

// Blocking
func (srv *Server) matchMaker() {
	for {
		p1 := <-srv.matcher
		p2 := <-srv.matcher

		a1 := p1.sendPing() == nil
		a2 := p2.sendPing() == nil

		if a1 && a2 {
			game := game.CreateGame(p1.player, p2.player)
			srv.addGame(game, [2]uuid.UUID{p1.id, p2.id})

			p1MatchStartMsg, err := MakeMessage(MatchStarted, &matchStarted{
				Char: p1.player.GetChar(),
				OpponentChar: p2.player.GetChar(),
			})
			
			p2MatchStartMsg, err2 := MakeMessage(MatchStarted, &matchStarted{
				Char: p2.player.GetChar(),
				OpponentChar: p1.player.GetChar(),
			})

			if err != nil || err2 != nil {
				log.Print("cannot make match started message")
			}

			p1.sendMessage(p1MatchStartMsg)
			p2.sendMessage(p2MatchStartMsg)
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
				srv.DeleteConnection(conn.id)
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

	if err != nil {
		return err
	}

	if response.Approved {
		if room := srv.GetRoomWithConnectionID(conn.id); room != nil {
			idx := slices.Index(room.connections[:], conn.id)
			opponentID := room.connections[(idx+1)%2]
		
			msgForOpponent, err := MakeMessage(OpponentMove, &moveMessage{
				X: pos.X,
				Y: pos.Y,
			})
			if err != nil {
				return err
			}

			srv.connections[opponentID].sendMessage(msgForOpponent)
		} else {
			return errors.New("cannot find room with player")
		}
	}

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

func (srv *Server) GetRoomWithConnectionID(id uuid.UUID) *room {
	idx := slices.IndexFunc(srv.rooms, func(r *room) bool {
		return slices.Contains(r.connections[:], id)
	})
	if idx == -1 {
		return nil
	}
	return srv.rooms[idx]
}