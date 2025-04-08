package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"GridPlay/assert"
	"GridPlay/gameServer"

	"github.com/lmittmann/tint"
)

var srv *gameServer.Server
var loopStopSignal chan bool
var isLoopRunning bool

func handleConnections(w http.ResponseWriter, r *http.Request) {
	assert.NotNil(srv, "server was nil")

	err := srv.HandleConnection(w, r)

	if err != nil {
		slog.Error("cannot add connection", "error", err)
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
	assert.NotNil(srv, "server was nil")

	go loop()
	srv.StartLoop()
}

func stopLoop() {
	assert.Assert(isLoopRunning, "loop wasn't running")
	assert.NotNil(srv, "server was nil")

	loopStopSignal <- true
	srv.StopLoop()
}

func main() {
	assertFile, err := os.Create("assert.txt")
	assert.NoError(err, "unable to open assert file")
	assert.ToWriter(assertFile)

	w := os.Stderr

	slog.SetDefault(slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}), 
	))

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