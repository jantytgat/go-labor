package labor

type EventTriggerer interface {
	HasEvent() bool
	Event() Event
}

type Envelope struct {
	Sender   string
	Receiver string
	Message  any
}

func (e *Envelope) HasEvent() bool {
	_, ok := e.Message.(EventTriggerer)
	return ok
}

func (e *Envelope) Event() Event {
	event, ok := e.Message.(EventTriggerer)
	if ok {
		return event.Event()
	}
	return ""
}
