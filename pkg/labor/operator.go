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
	operatorStartedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator started"}
	operatorStoppedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator stopped"}
)

type operatorConfig struct {
	Address *Address
	Router  *router
}

func newOperator(config operatorConfig) *operator {
	return &operator{
		config: config,
	}
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
	fmt.Println("Received envelope:", e)
}

func (o *operator) Start(ctx context.Context) {
	o.ctx, o.ctxCancel = context.WithCancel(ctx)

	defer o.config.Router.Send(Envelope{
		Sender:  o,
		Message: operatorStartedEvent,
	})
}

func (o *operator) Stop() {
	if o.ctxCancel != nil {
		defer o.config.Router.Send(Envelope{
			Sender:  o,
			Message: operatorStoppedEvent,
		})

		o.ctxCancel()
	}
}
