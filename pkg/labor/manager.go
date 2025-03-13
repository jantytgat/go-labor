package labor

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
)

const (
	managerKind Kind = "manager"
	managerId        = "root"
)

var (
	managerEnabledEvent         = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager enabled"}
	managerDisabledEvent        = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager disabled"}
	registerEvent               = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "register address"}
	unsupportedMessageEvent     = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "unsupported message"}
	managerReceivedJobEvent     = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager received job"}
	managerReceivedProcessEvent = Event{Category: laborEventCategory, Type: managerKind.String(), Message: "manager received process"}
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
		//o                   []*operator
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
		registry:           make(map[string]Addressable),
		broadcastListeners: make([]Addressable, 0),
		availableOperator:  chAvailableOperator,
		eventLogger:        l,
		eventLogLevel:      c.EventLogLevel,
	}

	//o = make([]*operator, c.MaxOperators)
	for i := 0; i < c.MaxOperators; i++ {
		m.Register(newOperator(fmt.Sprintf("operator_%d", i+1), m, chAvailableOperator), false)
	}

	//m.operators = o
	return m
}

type Manager struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	config    ManagerConfig
	address   *Address
	enabled   bool
	//operators         []*operator
	eventLogger        *slog.Logger
	eventLogLevel      slog.Level
	availableOperator  chan Addressable
	registry           map[string]Addressable
	broadcastListeners []Addressable
	mux                sync.RWMutex
}

func (m *Manager) Address() *Address {
	return m.address
}

func (m *Manager) broadcast(e envelope) {
	m.mux.RLock()
	defer m.mux.RUnlock()
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

func (m *Manager) handleProcess(e envelope) {
	if process, ok := e.Message.(Process); ok {
		m.logEvent(e.ctx, m, managerReceivedProcessEvent.WithInfo(process))

		var processor Addressable
		var err error
		if processor, err = m.processor(process.Task); err != nil {
			_ = NewProcessor(m, process.Task)
			processor, err = m.processor(process.Task)
			if err != nil {
				panic(err)
			}
		}

		m.logEvent(e.ctx, m, Event{
			Category: "temp",
			Type:     "temp",
			Message:  "send to processor",
			Info:     nil,
		})

		fmt.Println("available processor", processor)
		processor.Receive(envelope{
			ctx:      e.ctx,
			Sender:   e.Sender,
			Receiver: processor,
			Message:  process,
		})
	}
}

func (m *Manager) handleJob(e envelope) {
	if request, ok := e.Message.(Job); ok {
		m.logEvent(e.ctx, m, managerReceivedJobEvent.WithInfo(request.Name))

		availableOperator := <-m.availableOperator

		m.send(envelope{
			ctx:      e.ctx,
			Sender:   e.Sender,
			Receiver: availableOperator,
			Message:  e.Message,
		})
	}
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

func (m *Manager) processor(t Task) (Addressable, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()

	address := m.address.Child(processKind, reflect.TypeOf(t).String())
	fmt.Println(address.String())
	var processor Addressable
	var ok bool

	if processor, ok = m.registry[address.String()]; ok {
		fmt.Println("processor found")
		p := processor.(*Processor)
		pa := <-p.chAvailable
		return pa, nil
	}
	fmt.Println("processor not found")
	return nil, fmt.Errorf("unknown processor: %s", address)
}

func (m *Manager) Receive(e envelope) {
	switch e.Message.(type) {
	case Job:
		m.handleJob(e)
	case Process:
		m.handleProcess(e)
	default:
		m.logEvent(e.ctx, m, unsupportedMessageEvent)
	}
}

func (m *Manager) Register(a Addressable, broadcast bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.logEvent(context.TODO(), m, registerEvent.WithInfo(a.Address().String()))
	m.registry[a.Address().String()] = a

	if broadcast {
		m.broadcastListeners = append(m.broadcastListeners, a)
	}
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
