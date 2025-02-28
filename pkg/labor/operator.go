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
// operatorStartedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator started"}
// operatorStoppedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator stopped"}
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
			msg := Event{
				Category: laborEventCategory,
				Type:     operatorKind.String(),
				Message:  "received job",
				Info:     request.Name,
			}
			o.config.Router.Send(Envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: nil,
				Message:  msg,
			})

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
}

//
//func (o *operator) Start(ctx context.Context) {
//	o.ctx, o.ctxCancel = context.WithCancel(ctx)
//
//	defer o.config.Router.Send(Envelope{
//		Sender:  o,
//		Message: operatorStartedEvent,
//	})
//}
//
//func (o *operator) Stop() {
//	if o.ctxCancel != nil {
//		defer o.config.Router.Send(Envelope{
//			Sender:  o,
//			Message: operatorStoppedEvent,
//		})
//
//		o.ctxCancel()
//	}
//}
