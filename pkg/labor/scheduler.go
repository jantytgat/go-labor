package labor

import (
	"context"
	"log/slog"
)

const (
	schedulerKind Kind   = "scheduler"
	schedulerId   string = "root"
)

var (
	schedulerUnsupportedMessageEvent = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "unsupported message"}
	schedulerReceivedJobEvent        = Event{Category: laborEventCategory, Type: schedulerKind.String(), Message: "scheduler received job"}
)

type schedulerConfig struct {
	Address           *Address
	Router            *router
	AvailableOperator chan Addressable
	EventLogger       *slog.Logger
	EventLogLevel     slog.Level
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

func (s *scheduler) Receive(e envelope) {
	switch e.Message.(type) {
	case Request:
		if request, ok := e.Message.(Request); ok {
			s.logEvent(e.ctx, s, schedulerReceivedJobEvent.WithInfo(request.Name))

			availableOperator := <-s.config.AvailableOperator

			s.config.Router.Send(envelope{
				ctx:      e.ctx,
				Sender:   e.Sender,
				Receiver: availableOperator,
				Message:  e.Message,
			})
		}
	default:
		s.logEvent(e.ctx, s, schedulerUnsupportedMessageEvent)
	}
}

func (s *scheduler) logEvent(ctx context.Context, sender Addressable, event Event) {
	s.config.EventLogger.LogAttrs(
		ctx,
		s.config.EventLogLevel,
		event.String(),
		event.LogValue(sender.Address()))
}
