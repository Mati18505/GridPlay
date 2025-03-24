package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"TicTacToe/gameServer"
)

var srv *gameServer.Server

func handleConnections(w http.ResponseWriter, r *http.Request) {
    socket, err := upgrader.Upgrade(w, r, nil)
	
    if err != nil {
		log.Println(err)
		r.Body.Close() // Is it needed?
        return
    }

	conn := gameServer.CreateConnection(socket)
	err = srv.AddConnection(conn)
	if err != nil {
		log.Println("cannot add connection")
	}
}

func main() {
	srv = gameServer.InitGameServer()
	srv.StartLoop()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.HandleFunc("/ws", handleConnections)

	e := http.ListenAndServe(":4000", nil)

	srv.EndLoop()
	if e != nil {
		log.Fatal("ListenAndServe: ", e)
	}
}

var upgrader = websocket.Upgrader {
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}