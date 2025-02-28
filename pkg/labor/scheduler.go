package labor

import (
	"context"
)

const (
	schedulerKind Kind   = "scheduler"
	schedulerId   string = "root"
)

var (
	//schedulerStartedEvent            = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "scheduler started"}
	//schedulerStoppedEvent            = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "scheduler stopped"}
	schedulerUnsupportedMessageEvent = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "unsupported message"}
)

type schedulerConfig struct {
	Address           *Address
	Router            *router
	AvailableOperator chan Addressable
	Enabled           bool
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
	switch e.Message.(type) {
	case Request:
		if request, ok := e.Message.(Request); ok {
			msg := Event{
				Category: laborEventCategory,
				Type:     schedulerKind.String(),
				Message:  "received job",
				Info:     request.Name,
			}
			s.config.Router.Send(Envelope{
				ctx:      e.ctx,
				Sender:   s,
				Receiver: nil,
				Message:  msg,
			})

			availableOperator := <-s.config.AvailableOperator

			s.config.Router.Send(Envelope{
				ctx:      e.ctx,
				Sender:   e.Sender,
				Receiver: availableOperator,
				Message:  e.Message,
			})
		}
	default:
		s.config.Router.Send(Envelope{
			ctx:      e.ctx,
			Sender:   s,
			Receiver: nil,
			Message:  schedulerUnsupportedMessageEvent,
		})
	}
}

//
//func (s *scheduler) Start(ctx context.Context) {
//	s.ctx, s.ctxCancel = context.WithCancel(ctx)
//
//	defer s.config.Router.Send(Envelope{
//		Sender:  s,
//		Message: schedulerStartedEvent,
//	})
//}
//
//func (s *scheduler) Stop() {
//	if s.ctxCancel != nil {
//		defer s.config.Router.Send(Envelope{
//			Sender:  s,
//			Message: schedulerStoppedEvent,
//		})
//
//		s.ctxCancel()
//	}
//}
