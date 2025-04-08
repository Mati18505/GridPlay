package server

import "TicTacToe/gameServer/internal/server/serverEvents"

type Mediator interface {
	Notify(e serverEvents.MediatorEvent)
}