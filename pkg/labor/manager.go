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

func NewManager(config ManagerConfig) *Manager {
	rConfig := routerConfig{
		address: config.Address.Child(routerKind, routerId),
		EventLogger: config.EventLogger.With(
			slog.Group(
				"manager",
				slog.Any("address", config.Address.LogValue()))),
		EventLogLevel: config.EventLogLevel,
	}
	r := newRouter(rConfig)

	chAvailableOperator := make(chan Addressable, config.MaxOperators)

	sConfig := schedulerConfig{
		Router:            r,
		Address:           config.Address.Child(schedulerKind, schedulerId),
		AvailableOperator: chAvailableOperator,
		Enabled:           config.EnableScheduler,
	}

	operators := make([]*operator, config.MaxOperators)
	for i := 0; i < config.MaxOperators; i++ {
		oConfig := operatorConfig{
			Router:            r,
			Address:           config.Address.Child(operatorKind, fmt.Sprintf("operator_%d", i)),
			AvailableOperator: chAvailableOperator,
			Enabled:           config.EnableOperator,
		}
		operators[i] = newOperator(oConfig)
	}

	return &Manager{
		config:    config,
		scheduler: newScheduler(sConfig),
		operator:  operators,
		router:    r,
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

func (m *Manager) Enabled() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.enabled
}

func (m *Manager) Receive(e Envelope) {
	m.router.Send(Envelope{
		Sender:  m,
		Message: managerReceivedMessageEvent,
	})
}

func (m *Manager) Start(ctx context.Context) {
	defer func() {
		m.router.Send(Envelope{
			Sender:  m,
			Message: managerStartedEvent,
		})
		m.enable()
	}()

	m.ctx, m.ctxCancel = context.WithCancel(ctx)

	m.router.enable()
	//if m.config.EnableScheduler {
	//	m.scheduler.Start(m.ctx)
	//}

	go m.checkPoison()
}

func (m *Manager) Stop() {
	defer m.router.Send(Envelope{
		Sender:  m,
		Message: managerStoppedEvent,
	})
	m.disable()
	//m.scheduler.Stop()
	m.router.disable()
	m.ctxCancel()
}

func (m *Manager) checkPoison() {
	for {
		select {
		case <-m.ctx.Done():
			m.Stop()
			return
		}
	}
}

func (m *Manager) disable() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.enabled = false
}

func (m *Manager) enable() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.enabled = true
}
