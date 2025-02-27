package labor

import (
	"context"
)

const (
	schedulerKind Kind   = "scheduler"
	schedulerId   string = "root"
)

var (
	schedulerStartedEvent = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "scheduler started"}
	schedulerStoppedEvent = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "scheduler stopped"}
)

type schedulerConfig struct {
	Address *Address
	Router  *router
	Enabled bool
}

func newScheduler(config schedulerConfig) *scheduler {
	s := &scheduler{
		config: config,
	}
	config.Router.Register(s)
	return s
}

type scheduler struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	config    schedulerConfig
}

func (s *scheduler) Address() *Address {
	return s.config.Address
}

func (s *scheduler) Receive(e Envelope) {
	if event, ok := e.Message.(Event); ok {
		msg := Event{
			Category: "labor",
			Type:     "scheduler",
			Message:  "received job",
			Info:     event.Info,
		}
		s.config.Router.Send(Envelope{
			ctx:      nil,
			Sender:   e.Receiver,
			Receiver: nil,
			Message:  msg,
		})
	}

}

func (s *scheduler) Start(ctx context.Context) {
	s.ctx, s.ctxCancel = context.WithCancel(ctx)

	defer s.config.Router.Send(Envelope{
		Sender:  s,
		Message: schedulerStartedEvent,
	})
}

func (s *scheduler) Stop() {
	if s.ctxCancel != nil {
		defer s.config.Router.Send(Envelope{
			Sender:  s,
			Message: schedulerStoppedEvent,
		})

		s.ctxCancel()
	}
}
