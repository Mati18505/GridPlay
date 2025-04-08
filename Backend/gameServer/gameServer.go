package gameServer

import (
	"TicTacToe/assert"
	"TicTacToe/gameServer/internal/connection"
	"TicTacToe/gameServer/internal/server/mediator"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	srvMediator *mediator.ServerMediator
}

func InitGameServer() *Server {
	srv := &Server{
		srvMediator: mediator.CreateServerMediator(),
	}

	return srv
}

func (srv *Server) HandleConnection(w http.ResponseWriter, r *http.Request) error {
	assert.NotNil(srv.srvMediator, "mediator was nil")

	slog.Debug("creating socket")

    socket, err := upgrader.Upgrade(w, r, nil)
	defer r.Body.Close() 

    if err != nil {
        return err
    }

	slog.Debug("adding socket as connection")
	conn := connection.CreateConnection(socket)
	srv.srvMediator.AddConnection(conn)

	return nil
}

var upgrader = websocket.Upgrader {
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (srv *Server) StartLoop() {
	assert.NotNil(srv.srvMediator, "mediator was nil")

	srv.srvMediator.StartLoop()
}

func (srv *Server) StopLoop() {
	assert.NotNil(srv.srvMediator, "mediator was nil")

	srv.srvMediator.StopLoop()
}

func (srv *Server) Update() {
	assert.NotNil(srv.srvMediator, "mediator was nil")

	srv.srvMediator.Update()
}