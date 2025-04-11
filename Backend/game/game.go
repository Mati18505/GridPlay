package game

import (
	"GridPlay/gameServer/externalEvent"
)

type MoveParam = string

type Game interface {
	GetGameStartMessage(playerId int) externalEvent.EventGameMessage
	GetWinState() WinState 
	HandleGameMsg(eGameMsg externalEvent.EventGameMessage) ([]externalEvent.EventGameMessage, error)
}

type Player interface {
	GetId() int
}