package labor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type ManagerConfig struct {
	Address       *Address
	MaxOperators  int
	EventLogger   *slog.Logger
	EventLogLevel slog.Level
}

func NewManager(c ManagerConfig) *Manager {
	var (
		l *slog.Logger
		m *Manager
	)

	l = c.EventLogger.With(
		slog.Group(
			"manager",
			slog.Any("address", c.Address.LogValue())))

	m = &Manager{
		config:             c,
		address:            c.Address,
		enabled:            false,
		registry:           make(map[string]Addressable),
		broadcastListeners: make([]Addressable, 0),
		chOperator:         make(chan Addressable, c.MaxOperators),
		eventLogger:        l,
		eventLogLevel:      c.EventLogLevel,
	}

	for i := 0; i < c.MaxOperators; i++ {
		newOperator(fmt.Sprintf("operator_%d", i+1), m)
	}

	return m
}

type Manager struct {
	ctx                context.Context
	ctxCancel          context.CancelFunc
	config             ManagerConfig
	address            *Address
	enabled            bool
	eventLogger        *slog.Logger
	eventLogLevel      slog.Level
	chOperator         chan Addressable
	registry           map[string]Addressable
	broadcastListeners []Addressable
	mux                sync.Mutex
}

func (m *Manager) Address() *Address {
	return m.address
}

func (m *Manager) broadcast(e envelope) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, broadcast := range m.broadcastListeners {
		if broadcast != nil {
			broadcast.Receive(e)
		}
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

func (m *Manager) Disable() {
	if m.IsEnabled() {
		m.disable()
		m.logEvent(m.ctx, m, managerDisabledEvent.WithInfo("disabled"))
	}
}

func (m *Manager) disable() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.enabled = false
}

func (m *Manager) Enable(ctx context.Context) {
	m.ctx, m.ctxCancel = context.WithCancel(ctx)
	go m.checkPoison()

	m.enable()
	m.logEvent(m.ctx, m, managerEnabledEvent.WithInfo("enabled"))
}

func (m *Manager) enable() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.enabled = true
}

func (m *Manager) handleJob(e envelope) {
	if job, ok := e.Message.(Job); ok {
		m.logEvent(e.ctx, m, managerReceivedJobEvent.WithInfo(job.Name))
		availableOperator := <-m.chOperator

		availableOperator.Receive(envelope{
			ctx:      e.ctx,
			Sender:   e.Sender,
			Receiver: availableOperator,
			Message:  e.Message,
		})
	}
}

func (m *Manager) IsEnabled() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.enabled
}

func (m *Manager) logEvent(ctx context.Context, sender Addressable, event Event) {
	// if m.eventLogLevel >= event.Level {
	m.eventLogger.LogAttrs(
		ctx,
		event.Level,
		event.String(),
		event.LogValue(sender.Address()))
	// }
}

func (m *Manager) processor(h Handler) Addressable {
	m.mux.Lock()
	address := m.address.Child(processKind, h.Name).String()
	_, ok := m.registry[address]
	m.mux.Unlock()

	if !ok {
		newProcessor(m, h)
	}

	// Only return the address of a processor when it is available
	return <-m.registry[address].(*processor).chAvailable
}

func (m *Manager) Receive(e envelope) {
	switch e.Message.(type) {
	case Job:
		m.handleJob(e)
	default:
		m.logEvent(e.ctx, m, unsupportedMessageEvent)
	}
}

func (m *Manager) Register(a Addressable, broadcast bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if _, ok := m.registry[a.Address().String()]; ok {
		return
	}
	m.registry[a.Address().String()] = a

	if broadcast {
		m.broadcastListeners = append(m.broadcastListeners, a)
	}
	m.logEvent(context.TODO(), m, registeredEvent.WithInfo(a.Address().String()))

}

func (m *Manager) Unregister(a Addressable) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.registry, a.Address().String())
}

func (m *Manager) send(e envelope) {
	if event, ok := e.Message.(Event); ok {
		m.logEvent(e.ctx, e.Sender, event)
	}

	if e.Receiver == nil {
		return
	}

	switch e.Receiver.Address().IsBroadcast() {
	case true:
		go m.broadcast(e)
	case false:
		e.Receiver.Receive(e)
	}
}
