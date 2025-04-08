package handlers

import (
	"GridPlay/assert"
	"GridPlay/gameServer/internal/event"
)

type Synchronizer struct {
	nextHandler Handler
	syncChannel chan event.Event
}

func CreateSynchronizer(nextHandler Handler) *Synchronizer {
	assert.NotNil(nextHandler, "next handler was nil")

	return &Synchronizer{
		nextHandler: nextHandler,
		syncChannel: make(chan event.Event, 256),
	}
}

func (sync *Synchronizer) Handle(e event.Event) {
	assert.NotNil(sync.syncChannel, "sync channel was nil")

	sync.syncChannel <- e
}

func (sync *Synchronizer) SyncTransferAll() {
	assert.NotNil(sync.syncChannel, "sync channel was nil")

	for {
		select {
		case e := <-sync.syncChannel:
			sync.sendToNextHandler(e)
		default:
			assert.Assert(sync.empty(), "SyncTransferAll should clear syncChannel")
			return
		}
	}
}

func (sync *Synchronizer) empty() bool {
	return len(sync.syncChannel) == 0
}

func (sync *Synchronizer) sendToNextHandler(e event.Event) {
	assert.NotNil(sync.nextHandler, "player next handler was nil")

	sync.nextHandler.Handle(e)
}