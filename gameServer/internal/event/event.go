package event

type Event interface {
	GetType() EventType
}

type EventType int
const (
	EventTypeExit EventType = iota
	EventTypeMove
	EventTypeSendMessage
)