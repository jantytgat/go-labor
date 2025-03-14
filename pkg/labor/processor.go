package labor

import (
	"context"
	"time"
)

const (
	processKind Kind = "process"
)

func newProcessor(m *Manager, h Handler) {
	var maxConcurrent int

	if maxConcurrent = h.MaxConcurrent; maxConcurrent == 0 {
		maxConcurrent = m.config.MaxOperators
	}

	p := &processor{
		address:     m.Address().Child(processKind, h.Name),
		execute:     h.Execute,
		timeout:     h.Timeout,
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
	execute     func(ctx context.Context, p Pipeline) Pipeline
	timeout     time.Duration
	chAvailable chan Addressable
}

func (p *processor) Address() *Address {
	return p.address
}

func (p *processor) handleProcess(e envelope) {
	p.manager.logEvent(e.ctx, p, processorHandleProcessEvent)

	pipeline, ok := e.Message.(Pipeline)
	if !ok {
		panic("invalid process message")
	}

	execCtx, cancel := context.WithTimeout(e.ctx, p.timeout)
	defer cancel()

	e.Sender.Receive(envelope{
		ctx:     e.ctx,
		Sender:  p,
		Message: p.execute(execCtx, pipeline),
	})
}

func (p *processor) makeAvailable(ctx context.Context) {
	p.manager.logEvent(ctx, p, processorAvailableEvent.WithInfo(p.address.id))
	p.chAvailable <- p
}

func (p *processor) Receive(e envelope) {
	defer p.makeAvailable(e.ctx)

	switch e.Message.(type) {
	case Pipeline:
		p.handleProcess(e)
	default:
		p.manager.logEvent(e.ctx, p, unsupportedMessageEvent)
	}
}
