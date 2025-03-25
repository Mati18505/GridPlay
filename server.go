package main

import (
	"log"
	"net/http"

	"TicTacToe/gameServer"
)

var srv *gameServer.Server

func handleConnections(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleConnection(w, r)

	if err != nil {
		log.Println("cannot add connection")
	}
}

func main() {
	srv = gameServer.InitGameServer()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.HandleFunc("/ws", handleConnections)

	e := http.ListenAndServe(":4000", nil)

	if e != nil {
		log.Fatal("ListenAndServe: ", e)
	}
}