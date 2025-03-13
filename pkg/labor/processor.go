package labor

import (
	"context"
	"reflect"
	"time"
)

const (
	processKind Kind = "process"
)

var (
	processorHandleProcessEvent = Event{Category: laborEventCategory, Type: processKind.String(), Message: "handle process"}
)

func NewProcessor(m *Manager, t Task) *Processor {
	p := &Processor{
		address:     m.Address().Child(processKind, reflect.TypeOf(t).String()),
		execute:     t.Handler().execute,
		timeout:     t.Handler().timeout,
		chAvailable: make(chan Addressable, t.Handler().maxConcurrent),
	}

	m.Register(p, false)

	for i := 0; i < t.Handler().maxConcurrent; i++ {
		p.chAvailable <- p
	}

	return p
}

type Processor struct {
	address     *Address
	manager     *Manager
	execute     func(ctx context.Context, t Task, data any)
	timeout     time.Duration
	chAvailable chan Addressable
}

func (p *Processor) Address() *Address {
	return p.address
}

func (p *Processor) Available() chan Addressable {
	return p.chAvailable
}

func (p *Processor) Receive(e envelope) {
	defer p.makeAvailable()

	switch e.Message.(type) {
	case Process:
		process, ok := e.Message.(Process)
		if !ok {
			// Error handling??
			return
		}

		p.manager.logEvent(e.ctx, p, processorHandleProcessEvent.WithInfo(process))

		execCtx, cancel := context.WithTimeout(e.ctx, p.timeout)
		defer cancel()

		p.execute(execCtx, process.Task, process.Data)

		p.manager.Receive(envelope{
			ctx:      e.ctx,
			Sender:   p,
			Receiver: e.Sender,
			Message:  process,
		})
	default:
		p.manager.logEvent(e.ctx, p, UnsupportedMessageEvent)
	}
}

func (p *Processor) makeAvailable() {
	p.chAvailable <- p
}
