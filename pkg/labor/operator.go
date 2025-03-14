package labor

import (
	"context"
	"fmt"
	"sync"
)

func newOperator(name string, m *Manager) {
	o := &operator{
		address: m.address.Child(operatorKind, name),
		manager: m,
	}

	m.Register(o, false)
	o.makeAvailable(m.ctx)
}

type operator struct {
	address    *Address
	manager    *Manager
	chPipeline chan Pipeline
	ctx        context.Context
	ctxCancel  context.CancelFunc
	mux        sync.Mutex
}

func (o *operator) Address() *Address {
	return o.address
}

func (o *operator) createProcessChannel(size int) {
	o.mux.Lock()
	defer o.mux.Unlock()
	if o.chPipeline != nil {
		panic("channel already defined")
	}
	o.chPipeline = make(chan Pipeline, size)
}

func (o *operator) handleJob(e envelope) {
	job, ok := e.Message.(Job)
	if !ok {
		return
	}
	o.manager.logEvent(e.ctx, o, operatorReceivedJobEvent.WithInfo(job.Name))

	if len(job.Sequence) == 0 {
		o.manager.logEvent(e.ctx, o, emptySequenceEvent.WithInfo(job.Name))
		return
	}

	o.createProcessChannel(len(job.Sequence))

	o.manager.logEvent(e.ctx, o, operatorSequenceStartEvent.WithInfo(fmt.Sprintf("%s", job.Name)))

	for i, process := range job.Sequence {
		o.manager.logEvent(e.ctx, o, operatorProcessStartEvent.WithInfo(fmt.Sprintf("%s_#step_%d", job.Name, i)))
		o.handleProcess(e.ctx, process, job.Data)
		_ = <-o.chPipeline
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

func (o *operator) handleProcess(ctx context.Context, process Process, data Pipeline) {
	o.manager.processor(process.Handler).Receive(envelope{
		ctx:     ctx,
		Sender:  o,
		Message: data,
	})
}

func (o *operator) makeAvailable(ctx context.Context) {
	o.resetProcessChannel()
	o.manager.logEvent(ctx, o, operatorAvailableEvent.WithInfo(o.Address().id))
	o.manager.chOperator <- o
}

func (o *operator) Receive(e envelope) {
	switch e.Message.(type) {
	case Pipeline:
		o.chPipeline <- e.Message.(Pipeline)
	case Job:
		o.handleJob(e)
	default:
		o.manager.logEvent(e.ctx, o, unsupportedMessageEvent)
	}
}

func (o *operator) resetProcessChannel() {
	o.mux.Lock()
	defer o.mux.Unlock()

	if o.chPipeline != nil {
		close(o.chPipeline)
		o.chPipeline = nil
	}
}
