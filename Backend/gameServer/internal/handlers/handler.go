package handlers

import "GridPlay/gameServer/internal/event"

type Handler interface {
	Handle(e event.Event)
};