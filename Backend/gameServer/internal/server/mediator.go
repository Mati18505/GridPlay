package server

import "GridPlay/gameServer/internal/server/serverEvents"

type Mediator interface {
	Notify(e serverEvents.MediatorEvent)
}