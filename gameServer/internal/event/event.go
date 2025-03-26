package event

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
	EventTypeExit
	EventTypeMove
	EventTypeSendMessage
)