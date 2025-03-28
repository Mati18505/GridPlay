package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"TicTacToe/assert"
	"TicTacToe/gameServer"
)

var srv *gameServer.Server
var loopStopSignal chan bool
var isLoopRunning bool

func handleConnections(w http.ResponseWriter, r *http.Request) {
	assert.NotNil(srv, "server was nil")

	err := srv.HandleConnection(w, r)

	if err != nil {
		log.Printf("cannot add connection, error: %s", err)
	}
}

func loop() {
	assert.NotNil(srv, "server was nil")

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
	assert.Assert(!isLoopRunning, "loop was already running")

	go loop()
}

func stopLoop() {
	assert.Assert(isLoopRunning, "loop wasn't running")

	loopStopSignal <- true
}

func main() {
	assertFile, err := os.Create("assert.txt")
	assert.NoError(err, "unable to open assert file")
	assert.ToWriter(assertFile)

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