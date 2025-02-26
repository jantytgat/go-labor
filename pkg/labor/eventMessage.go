package labor

type EventMessage struct {
	event Event
}

func (e EventMessage) HasEvent() bool {
	return true
}

func (e EventMessage) Event() Event {
	return e.event
}
