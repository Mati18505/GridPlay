package handlers

import (
	"GridPlay/assert"
	"GridPlay/gameServer/internal/event"
	"GridPlay/gameServer/internal/server"
	"GridPlay/gameServer/internal/server/serverEvents"
)

type ServerHandler struct {
	mediator server.Mediator
	sync *Synchronizer
}

func CreateServerHandler(mediator server.Mediator) *ServerHandler {
	assert.NotNil(mediator, "mediator was nil")

	srvHandler := &ServerHandler{
		mediator: mediator,
	}

	srvHandler.sync = CreateSynchronizer(srvHandler)

	return srvHandler
}

func (srvHandler *ServerHandler) Handle(e event.Event) {
	srvHandler.mediator.Notify(serverEvents.MediatorEvent{
		Sender: serverEvents.ServerHandler,
		Event: e,
	})
}

func (srvHandler *ServerHandler) GetSync() *Synchronizer {
	assert.NotNil(srvHandler.sync, "server handler synchronizer was nil")

	return srvHandler.sync
}