package matchmaker

import (
	"TicTacToe/gameServer/internal/event"

	"github.com/google/uuid"
)

type EventPlayersMatched struct {
	Ids [2]uuid.UUID
}

func (e EventPlayersMatched) GetType() event.EventType {
	return event.EventTypePlayersMatched
}