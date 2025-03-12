package labor

import (
	"context"
	"fmt"
)

const (
	operatorKind Kind = "operators"
)

var (
	operatorAvailableEvent   = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operators available"}
	operatorReceivedJobEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operators received job"}
)

type operatorConfig struct {
	Address           *Address
	Manager           *Manager
	AvailableOperator chan Addressable
}

func newOperator(config operatorConfig) *operator {
	o := &operator{
		config: config,
	}
	o.config.Manager.Register(o)
	o.config.Manager.logEvent(context.TODO(), o, operatorAvailableEvent.WithInfo(o.Address().id))
	config.AvailableOperator <- o
	return o
}

type operator struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	config    operatorConfig
}

func (o *operator) Address() *Address {
	return o.config.Address
}

func (o *operator) Receive(e envelope) {
	switch e.Message.(type) {
	case Request:
		if request, ok := e.Message.(Request); ok {
			o.config.Manager.logEvent(e.ctx, o, operatorReceivedJobEvent.WithInfo(request.Name))

			// Execute job

			// Send result back to the customer
			o.config.Manager.send(envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: e.Sender,
				Message:  fmt.Sprintf("completed job: %s by %s", request.Name, o.Address().String()),
			})
		}
	default:
		o.config.Manager.logEvent(e.ctx, o, UnsupportedMessageEvent)
	}
	o.config.Manager.logEvent(e.ctx, o, operatorAvailableEvent.WithInfo(o.Address().id))
	o.config.AvailableOperator <- o
}
