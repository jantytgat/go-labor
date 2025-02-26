package labor

import "context"

const (
	operatorKind Kind = "operator"
	operatorId        = "root"
)

var (
	operatorStartedEvent = Event{laborEventCategory, operatorKind.String(), "operator started"}
	operatorStoppedEvent = Event{laborEventCategory, operatorKind.String(), "operator stopped"}
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

func (o *operator) Start(ctx context.Context) {
	o.ctx, o.ctxCancel = context.WithCancel(ctx)

	defer o.config.Router.Process(Envelope{
		Sender:  o.Address(),
		Message: operatorStartedEvent,
	})
}

func (o *operator) Stop() {
	if o.ctxCancel != nil {
		defer o.config.Router.Process(Envelope{
			Sender:  o.Address(),
			Message: operatorStoppedEvent,
		})

		o.ctxCancel()
	}
}
