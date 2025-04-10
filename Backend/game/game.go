package game

import (
	"GridPlay/gameServer/externalEvent"
)

type MoveParam = string

type Game interface {
	GetWinState() WinState 
	HandleGameMsg(eGameMsg externalEvent.EventGameMessage) []externalEvent.EventGameMessage
}

type Player interface {
	GetId() int
}