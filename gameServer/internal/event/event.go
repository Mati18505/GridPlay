package event

type Event struct {
	eType EventType
}

func (e *Event) GetType() EventType {
	return e.eType
}

func CreateEvent(eType EventType) Event {
	return Event{
		eType: eType,
	}
}

type EventTypeEnum int
const (
	EventTypeExit EventTypeEnum = iota
	EventTypeMove
	EventTypeSendMessage
)

type EventType interface{
	GetEventType() EventTypeEnum
}