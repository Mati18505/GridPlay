package serverEvents

import (
	"GridPlay/gameServer/internal/event"
)

type MediatorEvent struct {
	Sender Sender
	Event event.Event
}

type Sender int
const (
	ServerHandler Sender = iota
	Matchmaker
)