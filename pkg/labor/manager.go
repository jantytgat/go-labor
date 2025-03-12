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
	managerStartedEvent     = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager started"}
	managerStoppedEvent     = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager stopped"}
	UnsupportedMessageEvent = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "unsupported message"}
	ReceivedJobEvent        = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "router received job"}
	ReceivedProcessEvent    = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "router received process"}
)

type ManagerConfig struct {
	Address       *Address
	MaxOperators  int
	EventLogger   *slog.Logger
	EventLogLevel slog.Level
}

func NewManager(c ManagerConfig) *Manager {
	var (
		l                   *slog.Logger
		m                   *Manager
		o                   []*operator
		chAvailableOperator chan Addressable
	)

	l = c.EventLogger.With(
		slog.Group(
			"manager",
			slog.Any("address", c.Address.LogValue())))

	chAvailableOperator = make(chan Addressable, c.MaxOperators)

	m = &Manager{
		config:             c,
		address:            c.Address,
		enabled:            false,
		broadcastListeners: make(map[Addressable]bool),
		availableOperator:  chAvailableOperator,
		eventLogger:        l,
		eventLogLevel:      c.EventLogLevel,
	}

	o = make([]*operator, c.MaxOperators)
	for i := 0; i < c.MaxOperators; i++ {
		oConfig := operatorConfig{
			Manager:           m,
			Address:           c.Address.Child(operatorKind, fmt.Sprintf("operator_%d", i+1)),
			AvailableOperator: chAvailableOperator,
		}
		o[i] = newOperator(oConfig)
	}

	m.operators = o
	return m
}

type Manager struct {
	ctx                context.Context
	ctxCancel          context.CancelFunc
	config             ManagerConfig
	address            *Address
	enabled            bool
	operators          []*operator
	eventLogger        *slog.Logger
	eventLogLevel      slog.Level
	availableOperator  chan Addressable
	availableProcessor chan Addressable
	broadcastListeners map[Addressable]bool
	mux                sync.RWMutex
}

func (m *Manager) Address() *Address {
	return m.address
}

func (m *Manager) broadcast(e envelope) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	for contact, broadcast := range m.broadcastListeners {
		if !broadcast {
			continue
		}
		contact.Receive(e)
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
		m.logEvent(m.ctx, m, managerStoppedEvent.WithInfo("disabled"))
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
	m.logEvent(m.ctx, m, managerStartedEvent.WithInfo("enabled"))
}

func (m *Manager) enable() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.enabled = true
}

func (m *Manager) IsEnabled() bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.enabled
}

func (m *Manager) logEvent(ctx context.Context, sender Addressable, event Event) {
	m.eventLogger.LogAttrs(
		ctx,
		m.eventLogLevel,
		event.String(),
		event.LogValue(sender.Address()))
}

func (m *Manager) Receive(e envelope) {
	switch e.Message.(type) {
	case Request:
		if request, ok := e.Message.(Request); ok {
			m.logEvent(e.ctx, m, ReceivedJobEvent.WithInfo(request.Name))

			availableOperator := <-m.availableOperator

			m.send(envelope{
				ctx:      e.ctx,
				Sender:   e.Sender,
				Receiver: availableOperator,
				Message:  e.Message,
			})
		}
	case Process:
		if process, ok := e.Message.(Process); ok {
			m.logEvent(e.ctx, m, ReceivedProcessEvent.WithInfo(process.Name))

			availableProcessor := <-m.availableProcessor

			m.send(envelope{
				ctx:      e.ctx,
				Sender:   e.Sender,
				Receiver: availableProcessor,
				Message:  e.Message,
			})
		}
	default:
		m.logEvent(e.ctx, m, UnsupportedMessageEvent)
	}
}

func (m *Manager) Register(a Addressable) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.broadcastListeners[a] = true
}

func (m *Manager) Unregister(a Addressable) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.broadcastListeners, a)
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
