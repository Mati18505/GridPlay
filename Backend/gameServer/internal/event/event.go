package event

import "GridPlay/assert"

type Event interface {
	GetType() EventType
}

type EmptyEvent struct {}

func (emptyEvent EmptyEvent) GetType() EventType {
	return EventTypeNone;
}

type EventType int
const (
	EventTypeNone EventType = iota
	EventTypeDisconnect
	EventTypeRemoveRoom
	EventTypeGameMessage
	EventTypeSendMessage
	// server
	EventTypePlayersMatched
)

func (eType EventType) String() string {
	switch eType {
	case EventTypeNone:
		return "None"
	case EventTypeDisconnect:
		return "Disconnect"
	case EventTypeRemoveRoom:
		return "RemoveRoom"
	case EventTypeGameMessage:
		return "GameMessage"
	case EventTypeSendMessage:
		return "SendMessage"
	case EventTypePlayersMatched:
		return "PlayersMatched"
	default:
		assert.Never("unknown type of event", "server event", eType)
		return "Unknown"
	}
}