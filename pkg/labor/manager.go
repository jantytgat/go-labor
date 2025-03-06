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
	Address                *Address
	MaxOperators           int
	ManagerEventLogLevel   slog.Level
	RouterEventLogLevel    slog.Level
	SchedulerEventLogLevel slog.Level
	OperatorEventLogLevel  slog.Level
}

func NewManager(c ManagerConfig, l *slog.Logger) *Manager {
	var (
		r                   *router
		s                   *scheduler
		o                   []*operator
		p                   *processor
		chAvailableOperator chan Addressable
	)

	l = l.With(
		slog.Group(
			"manager",
			slog.Any("address", c.Address.LogValue())))

	chAvailableOperator = make(chan Addressable, c.MaxOperators)

	r = newRouter(routerConfig{
		address:       c.Address.Child(routerKind, routerId),
		EventLogger:   l,
		EventLogLevel: c.RouterEventLogLevel,
	})

	s = newScheduler(schedulerConfig{
		Router:            r,
		Address:           c.Address.Child(schedulerKind, schedulerId),
		AvailableOperator: chAvailableOperator,
		EventLogger:       l,
		EventLogLevel:     c.SchedulerEventLogLevel,
	})

	p = newProcessor(processorConfig{
		Router:  r,
		Address: c.Address.Child(processorKind, processorId),
	})

	o = make([]*operator, c.MaxOperators)
	for i := 0; i < c.MaxOperators; i++ {
		oConfig := operatorConfig{
			Router:            r,
			Address:           c.Address.Child(operatorKind, fmt.Sprintf("operator_%d", i+1)),
			AvailableOperator: chAvailableOperator,
			EventLogger:       l,
			EventLogLevel:     c.OperatorEventLogLevel,
		}
		o[i] = newOperator(oConfig)
	}

	return &Manager{
		config:    c,
		logger:    l,
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
	logger    *slog.Logger
	enabled   bool
	scheduler *scheduler
	operator  []*operator
	router    *router
	mux       sync.RWMutex
}

func (m *Manager) IsEnabled() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.enabled
}

func (m *Manager) Enable(ctx context.Context) {
	m.ctx, m.ctxCancel = context.WithCancel(ctx)
	go m.checkPoison()

	m.enable()
	m.logEvent(m.ctx, managerStartedEvent.WithInfo("enabled"))
}

func (m *Manager) Disable() {
	if m.IsEnabled() {
		m.disable()
		m.logEvent(m.ctx, managerStoppedEvent.WithInfo("disabled"))
	}
}

func (m *Manager) checkPoison() {
	for {
		select {
		case <-m.ctx.Done():
			m.Disable()
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

func (m *Manager) logEvent(ctx context.Context, event Event) {
	m.logger.LogAttrs(
		ctx,
		m.config.ManagerEventLogLevel,
		event.String(),
		event.LogValue(m.config.Address))
}
