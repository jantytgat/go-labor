package labor

import (
	"context"
	"fmt"
)

const (
	operatorKind Kind = "operator"
	operatorId        = "root"
)

var (
	operatorAvailableEvent   = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator available"}
	operatorReceivedJobEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator received job"}
)

type operatorConfig struct {
	Address           *Address
	Router            *router
	AvailableOperator chan Addressable
	Enabled           bool
}

func newOperator(config operatorConfig) *operator {
	o := &operator{
		config: config,
	}
	o.config.Router.Register(o)

	o.config.Router.Send(Envelope{
		ctx:      context.TODO(),
		Sender:   o,
		Receiver: nil,
		Message:  operatorAvailableEvent.WithInfo(o.Address().id),
	})
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

func (o *operator) Receive(e Envelope) {
	switch e.Message.(type) {
	case Request:
		if request, ok := e.Message.(Request); ok {
			// Event: operator received job
			o.config.Router.Send(Envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: nil,
				Message:  operatorReceivedJobEvent.WithInfo(request.Name),
			})

			// Send result back to the customer
			o.config.Router.Send(Envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: e.Sender,
				Message:  fmt.Sprintf("completed job: %s by %s", request.Name, o.Address().String()),
			})
		}
	default:
		o.config.Router.Send(Envelope{
			ctx:      e.ctx,
			Sender:   o,
			Receiver: nil,
			Message:  schedulerUnsupportedMessageEvent,
		})
	}

	o.config.AvailableOperator <- o
	o.config.Router.Send(Envelope{
		ctx:      e.ctx,
		Sender:   o,
		Receiver: nil,
		Message:  operatorAvailableEvent.WithInfo(o.Address().id),
	})
}

//
// func (o *operator) Start(ctx context.Context) {
//	o.ctx, o.ctxCancel = context.WithCancel(ctx)
//
//	defer o.config.Router.Send(Envelope{
//		Sender:  o,
//		Message: operatorStartedEvent,
//	})
// }
//
// func (o *operator) Stop() {
//	if o.ctxCancel != nil {
//		defer o.config.Router.Send(Envelope{
//			Sender:  o,
//			Message: operatorStoppedEvent,
//		})
//
//		o.ctxCancel()
//	}
// }
