package labor

import (
	"context"
	"fmt"
	"log/slog"
)

const (
	operatorKind Kind = "operator"
)

var (
	operatorAvailableEvent   = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator available"}
	operatorReceivedJobEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Message: "operator received job"}
)

type operatorConfig struct {
	Address           *Address
	Router            *router
	AvailableOperator chan Addressable
	EventLogger       *slog.Logger
	EventLogLevel     slog.Level
}

func newOperator(config operatorConfig) *operator {
	o := &operator{
		config: config,
	}
	o.config.Router.Register(o)
	o.logEvent(context.TODO(), o, operatorAvailableEvent.WithInfo(o.Address().id))
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
			o.logEvent(e.ctx, o, operatorReceivedJobEvent.WithInfo(request.Name))

			// Execute job

			// Send result back to the customer
			o.config.Router.Send(envelope{
				ctx:      e.ctx,
				Sender:   o,
				Receiver: e.Sender,
				Message:  fmt.Sprintf("completed job: %s by %s", request.Name, o.Address().String()),
			})
		}
	default:
		o.logEvent(e.ctx, o, schedulerUnsupportedMessageEvent)
	}
	o.logEvent(e.ctx, o, operatorAvailableEvent.WithInfo(o.Address().id))
	o.config.AvailableOperator <- o
}

func (o *operator) logEvent(ctx context.Context, sender Addressable, event Event) {
	o.config.EventLogger.LogAttrs(
		ctx,
		o.config.EventLogLevel,
		event.String(),
		event.LogValue(sender.Address()))
}
