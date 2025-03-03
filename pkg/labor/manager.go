package labor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

const (
	managerKind Kind = "manager"
	managerId        = "root"
)

var (
	managerStartedEvent         = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager started"}
	managerStoppedEvent         = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager stopped"}
	managerReceivedMessageEvent = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager received message"}
)

type ManagerConfig struct {
	Address         *Address
	EnableScheduler bool
	EnableOperator  bool
	EventLogger     *slog.Logger
	EventLogLevel   slog.Level
	MaxOperators    int
}

func NewManager(c ManagerConfig) *Manager {
	rConfig := routerConfig{
		address: c.Address.Child(routerKind, routerId),
		EventLogger: c.EventLogger.With(
			slog.Group(
				"manager",
				slog.Any("address", c.Address.LogValue()))),
		EventLogLevel: c.EventLogLevel,
	}
	r := newRouter(rConfig)

	chAvailableOperator := make(chan Addressable, c.MaxOperators)

	sConfig := schedulerConfig{
		Router:            r,
		Address:           c.Address.Child(schedulerKind, schedulerId),
		AvailableOperator: chAvailableOperator,
		Enabled:           c.EnableScheduler,
	}
	s := newScheduler(sConfig)

	o := make([]*operator, c.MaxOperators)
	for i := 0; i < c.MaxOperators; i++ {
		oConfig := operatorConfig{
			Router:            r,
			Address:           c.Address.Child(operatorKind, fmt.Sprintf("operator_%d", i+1)),
			AvailableOperator: chAvailableOperator,
			Enabled:           c.EnableOperator,
		}
		o[i] = newOperator(oConfig)
	}

	return &Manager{
		config:    c,
		scheduler: s,
		operator:  o,
		router:    r,
		enabled:   false,
	}
}

type Manager struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	config    ManagerConfig
	enabled   bool
	scheduler *scheduler
	operator  []*operator
	router    *router
	mux       sync.RWMutex
}

func (m *Manager) Address() *Address {
	return m.config.Address
}

func (m *Manager) IsEnabled() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.router.enabled
}

func (m *Manager) Receive(e Envelope) {
	m.router.Send(Envelope{
		Sender:  m,
		Message: managerReceivedMessageEvent.WithInfo(m.Address().id),
	})
}

func (m *Manager) Start(ctx context.Context) {
	var ctxPoison context.Context
	ctxPoison, m.ctxCancel = context.WithCancel(ctx)
	go m.checkPoison(ctxPoison)

	// Activate the router to accept messages
	m.router.enable()

	m.router.Send(Envelope{
		Sender:  m,
		Message: managerStartedEvent.WithInfo(m.Address().id),
	})
}

func (m *Manager) Stop() {
	// Stop the router
	m.router.disable() // TODO set timeout
	m.ctxCancel()
	m.router.Send(Envelope{
		Sender:  m,
		Message: managerStoppedEvent.WithInfo(m.Address().id),
	})
}

func (m *Manager) checkPoison(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			m.Stop()
			return
		}
	}
}
