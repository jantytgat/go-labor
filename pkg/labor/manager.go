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
	managerStartedEvent = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager started"}
	managerStoppedEvent = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager stopped"}
)

type ManagerConfig struct {
	Address         *Address
	EnableScheduler bool
	EnableOperator  bool
	EventLogger     *slog.Logger
	EventLogLevel   slog.Level
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

	sConfig := schedulerConfig{
		Router:  r,
		Address: config.Address.Child(schedulerKind, schedulerId),
	}

	oConfig := operatorConfig{
		Router:  r,
		Address: config.Address.Child(operatorKind, operatorId),
	}

	return &Manager{
		config:    config,
		scheduler: newScheduler(sConfig),
		operator:  newOperator(oConfig),
		router:    r,
	}
}

type Manager struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	config    ManagerConfig
	enabled   bool
	scheduler *scheduler
	operator  *operator
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

func (m *Manager) Start(ctx context.Context) {
	defer func() {
		m.router.Process(Envelope{
			Sender:  m.Address(),
			Message: managerStartedEvent,
		})
		m.enable()
	}()

	m.ctx, m.ctxCancel = context.WithCancel(ctx)

	m.router.Enable()
	if m.config.EnableScheduler {
		m.scheduler.Start(m.ctx)
	}

	if m.config.EnableOperator {
		m.operator.Start(m.ctx)
	}

	go m.checkPoison()
}

func (m *Manager) Stop() {
	defer m.router.Process(Envelope{
		Sender:  m.Address(),
		Message: managerStoppedEvent,
	})
	m.disable()
	m.scheduler.Stop()
	m.operator.Stop()
	m.router.Disable()
	m.ctxCancel()
}

func (m *Manager) AddJob(name string) error {
	if m.Enabled() {
		defer m.router.Process(Envelope{
			Sender: m.Address(),
			Message: Event{
				Category: "labor",
				Type:     "job",
				Message:  "job added",
				Info: struct {
					Name string
				}{Name: name},
			},
		})
		return nil
	}
	return fmt.Errorf("manager is stopped")
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
