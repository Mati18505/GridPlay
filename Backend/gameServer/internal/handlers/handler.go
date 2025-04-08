package handlers

import "TicTacToe/gameServer/internal/event"

type Handler interface {
	Handle(e event.Event)
};