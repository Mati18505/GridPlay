package room

import "GridPlay/gameServer/internal/handlers"

type state interface {
	handleDisconnect(playerId, opponentId int)
	handleMove(eMove handlers.EventMove)
}