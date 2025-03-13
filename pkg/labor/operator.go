package labor

import (
	"context"
	"fmt"
	"sync"
)

const (
	operatorKind Kind = "operators"
)

var (
	operatorAvailableEvent         = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator available"}
	operatorReceivedJobEvent       = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator received job"}
	operatorCompletedJobEvent      = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator completed job"}
	operatorSequenceStartEvent     = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator started sequence"}
	operatorSequenceCompletedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator completed sequence"}
	operatorProcessStartEvent      = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator started process"}
	operatorProcessRunningEvent    = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator process running"}
	operatorProcessCompletedEvent  = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator completed process"}
	emptySequenceEvent             = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "empty job sequence"}
)

func newOperator(name string, m *Manager, chAvailable chan Addressable) {
	o := &operator{
		address:     m.address.Child(processKind, name),
		manager:     m,
		chAvailable: chAvailable,
	}

	m.Register(o, false)
	o.makeAvailable(m.ctx)
	// m.logEvent(m.ctx, o, operatorAvailableEvent.WithInfo(o.address.id))
	// chAvailable <- o
}

type operator struct {
	ctx         context.Context
	ctxCancel   context.CancelFunc
	address     *Address
	manager     *Manager
	chAvailable chan Addressable
	chProcess   chan Process
	mux         sync.Mutex
}

func (o *operator) Address() *Address {
	return o.address
}

func (o *operator) Receive(e envelope) {
	switch e.Message.(type) {
	case Process:
		o.chProcess <- e.Message.(Process)
	case Job:
		o.handleJob(e)
	default:
		o.manager.logEvent(e.ctx, o, unsupportedMessageEvent)
	}
}

func (o *operator) handleJob(e envelope) {
	var job Job
	var ok bool
	if job, ok = e.Message.(Job); !ok {
		return
	}
	o.manager.logEvent(e.ctx, o, operatorReceivedJobEvent.WithInfo(job.Name))

	if len(job.Pipeline.Sequence) == 0 {
		o.manager.logEvent(e.ctx, o, emptySequenceEvent.WithInfo(job.Name))
		return
	}

	o.mux.Lock()
	if o.chProcess != nil {
		panic("channel already defined")
	}
	o.chProcess = make(chan Process, len(job.Pipeline.Sequence))
	o.mux.Unlock()

	o.manager.logEvent(e.ctx, o, operatorSequenceStartEvent.WithInfo(fmt.Sprintf("%s", job.Name)))
	for i, process := range job.Pipeline.Sequence {
		o.manager.logEvent(e.ctx, o, operatorProcessStartEvent.WithInfo(fmt.Sprintf("%s_#step_%d", job.Name, i)))

		env := envelope{
			ctx:     e.ctx,
			Sender:  o,
			Message: process,
		}
		// go o.manager.Receive(env)
		o.manager.handleProcess(env)
		o.manager.logEvent(e.ctx, o, operatorProcessRunningEvent.WithInfo(fmt.Sprintf("%s_#step_%d", job.Name, i)))
		_ = <-o.chProcess
		o.manager.logEvent(e.ctx, o, operatorProcessCompletedEvent.WithInfo(fmt.Sprintf("%s_#step_%d", job.Name, i)))
	}
	o.manager.logEvent(e.ctx, o, operatorSequenceCompletedEvent.WithInfo(fmt.Sprintf("%s", job.Name)))

	// Send result back to the customer
	e.Sender.Receive(envelope{
		ctx:     e.ctx,
		Sender:  o,
		Message: fmt.Sprintf("completed job: %s by %s", job.Name, o.Address().String()),
	})

	o.manager.logEvent(e.ctx, o, operatorCompletedJobEvent.WithInfo(job.Name))
	o.makeAvailable(e.ctx)
}

func (o *operator) makeAvailable(ctx context.Context) {
	o.mux.Lock()
	defer o.mux.Unlock()
	if o.chProcess != nil {
		close(o.chProcess)
		o.chProcess = nil
	}
	o.manager.logEvent(ctx, o, operatorAvailableEvent.WithInfo(o.Address().id))
	o.chAvailable <- o
}
