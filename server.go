package main

import (
	"log"
	"net/http"
	"time"

	"TicTacToe/gameServer"
)

var srv *gameServer.Server

func handleConnections(w http.ResponseWriter, r *http.Request) {
	err := srv.HandleConnection(w, r)

	if err != nil {
		log.Println("cannot add connection")
	}
}

var loopStopSignal chan bool

func loop() {
	for {
		select {
		case <- time.After(time.Millisecond * 50):
			srv.Update()
		case <- loopStopSignal:
			return
		}
	}
}

func startLoop() {
	go loop()
}

func stopLoop() {
	loopStopSignal <- true
}

func main() {
	srv = gameServer.InitGameServer()

	startLoop()
	defer stopLoop()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	http.HandleFunc("/ws", handleConnections)

	e := http.ListenAndServe(":4000", nil)

	if e != nil {
		log.Fatal("ListenAndServe: ", e)
	}
}