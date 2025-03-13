package labor

import (
	"context"
	"fmt"
)

const (
	operatorKind Kind = "operators"
)

var (
	operatorAvailableEvent        = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator available"}
	operatorReceivedJobEvent      = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator received job"}
	operatorCompletedJobEvent     = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator completed job"}
	operatorStartedProcessEvent   = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator started process"}
	operatorCompletedProcessEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator completed process"}
	emptySequenceEvent            = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "empty job sequence"}
)

func newOperator(name string, m *Manager, chAvailable chan Addressable) *operator {
	o := &operator{
		address:     m.address.Child(processKind, name),
		manager:     m,
		chAvailable: chAvailable,
	}

	m.Register(o, false)
	m.logEvent(context.TODO(), o, operatorAvailableEvent.WithInfo(o.Address().id))
	chAvailable <- o

	return o
}

type operator struct {
	ctx         context.Context
	ctxCancel   context.CancelFunc
	address     *Address
	manager     *Manager
	chAvailable chan Addressable
	chProcess   chan Process
}

func (o *operator) Address() *Address {
	return o.address
}

func (o *operator) Receive(e envelope) {
	switch e.Message.(type) {
	case Process:
		if process, ok := e.Message.(Process); ok {
			o.chProcess <- process
		}
	case Job:
		if job, ok := e.Message.(Job); ok {
			o.manager.logEvent(e.ctx, o, operatorReceivedJobEvent.WithInfo(job.Name))

			o.handleJob(e)

			// Send result back to the customer
			o.manager.send(envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: e.Sender, // customer address
				Message:  fmt.Sprintf("completed job: %s by %s", job.Name, o.Address().String()),
			})

			o.manager.logEvent(e.ctx, o, operatorCompletedJobEvent.WithInfo(job.Name))
		}
	default:
		o.manager.logEvent(e.ctx, o, unsupportedMessageEvent)
	}
	o.manager.logEvent(e.ctx, o, operatorAvailableEvent.WithInfo(o.Address().id))
	o.chAvailable <- o
}

func (o *operator) handleJob(e envelope) {
	var job Job
	var ok bool
	if job, ok = e.Message.(Job); !ok {
		return
	}

	if len(job.Pipeline.Sequence) == 0 {
		o.manager.logEvent(e.ctx, o, emptySequenceEvent.WithInfo(job.Name))
		return
	}
	for i, process := range job.Pipeline.Sequence {
		o.manager.logEvent(e.ctx, o, operatorStartedProcessEvent.WithInfo(fmt.Sprintf("%s_%d", job.Name, i)))

		o.manager.Receive(envelope{
			ctx:      e.ctx,
			Sender:   o,
			Receiver: o.manager,
			Message:  process,
		})

		process = <-o.chProcess
		o.manager.logEvent(e.ctx, o, operatorCompletedProcessEvent.WithInfo(fmt.Sprintf("%s_%d", job.Name, i)))
	}
}
