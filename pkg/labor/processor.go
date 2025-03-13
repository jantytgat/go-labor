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
	processorNotFoundEvent      = Event{Category: laborEventCategory, Type: processKind.String(), Message: "processor not found"}
	processorInitializedEvent   = Event{Category: laborEventCategory, Type: processKind.String(), Message: "processor initialized"}
	processorAvailableEvent     = Event{Category: laborEventCategory, Type: processKind.String(), Message: "processor available"}
	processorHandleProcessEvent = Event{Category: laborEventCategory, Type: processKind.String(), Message: "handle process"}
)

func newProcessor(m *Manager, t Task) {
	var maxConcurrent int
	if maxConcurrent = t.Handler().maxConcurrent; maxConcurrent == 0 {
		maxConcurrent = m.config.MaxOperators
	}

	p := &processor{
		address:     m.Address().Child(processKind, reflect.TypeOf(t).String()),
		execute:     t.Handler().execute,
		timeout:     t.Handler().timeout,
		chAvailable: make(chan Addressable, maxConcurrent),
		manager:     m,
	}

	m.Register(p, false)

	for i := 0; i < maxConcurrent; i++ {
		p.makeAvailable(m.ctx)
	}
}

type processor struct {
	address     *Address
	manager     *Manager
	execute     func(ctx context.Context, t Task, data any) Process
	timeout     time.Duration
	chAvailable chan Addressable
}

func (p *processor) Address() *Address {
	return p.address
}

func (p *processor) Receive(e envelope) {
	defer p.makeAvailable(e.ctx)
	p.manager.logEvent(e.ctx, p, processorHandleProcessEvent)

	switch e.Message.(type) {
	case Process:
		process, ok := e.Message.(Process)
		if !ok {
			// TODO ERROR HANDLING
		}

		execCtx, cancel := context.WithTimeout(e.ctx, p.timeout)
		defer cancel()

		res := p.execute(execCtx, process.Task, process.Data)
		env := envelope{
			ctx:     e.ctx,
			Sender:  p,
			Message: res,
		}
		e.Sender.Receive(env)
	default:
		p.manager.logEvent(e.ctx, p, unsupportedMessageEvent)
	}
}

func (p *processor) makeAvailable(ctx context.Context) {
	p.manager.logEvent(ctx, p, processorAvailableEvent.WithInfo(p.address.id))
	p.chAvailable <- p
}
