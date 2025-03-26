package handlers

import "TicTacToe/gameServer/internal/event"

type Synchronizer struct {
	nextHandler Handler
	syncChannel chan event.Event
}

func CreateSynchronizer(nextHandler Handler) *Synchronizer {
	return &Synchronizer{
		nextHandler: nextHandler,
		syncChannel: make(chan event.Event, 256),
	}
}

func (sync *Synchronizer) Handle(e event.Event) {
	sync.syncChannel <- e
}

func (sync *Synchronizer) SyncTransferAll() {
	for {
		select {
		case e := <-sync.syncChannel:
			sync.nextHandler.Handle(e)
		default:
			return
		}
	}
}