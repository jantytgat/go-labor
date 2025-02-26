package labor

import "context"

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

func (s *scheduler) Start(ctx context.Context) {
	s.ctx, s.ctxCancel = context.WithCancel(ctx)

	defer s.config.Router.Process(Envelope{
		Sender:  s.Address(),
		Message: schedulerStartedEvent,
	})
}

func (s *scheduler) Stop() {
	if s.ctxCancel != nil {
		defer s.config.Router.Process(Envelope{
			Sender:  s.Address(),
			Message: schedulerStoppedEvent,
		})

		s.ctxCancel()
	}
}
