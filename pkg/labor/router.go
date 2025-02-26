package labor

import (
	"context"
	"log/slog"
	"sync"
)

const (
	routerKind Kind   = "router"
	routerId   string = "root"
)

type routerConfig struct {
	address       *Address
	EventLogger   *slog.Logger
	EventLogLevel slog.Level
}

func newRouter(config routerConfig) *router {
	return &router{
		config:   config,
		contacts: make(map[*Address]bool),
	}
}

type router struct {
	config      routerConfig
	enabled     bool
	contacts    map[*Address]bool
	eventLogger *slog.Logger
	mux         sync.RWMutex
}

func (r *router) Broadcast(e Envelope) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	for contact, broadcast := range r.contacts {
		if !broadcast {
			continue
		}
		e.Receiver = contact
		r.Send(e)
	}
}

func (r *router) Disable() {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.enabled = false
}

func (r *router) Enable() {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.enabled = true
}

func (r *router) LogEvent(ctx context.Context, sender *Address, event Event) {
	r.config.EventLogger.LogAttrs(
		ctx,
		r.config.EventLogLevel,
		event.String(),
		event.LogValue(sender))
}

func (r *router) Process(e Envelope) {
	if event, ok := e.Message.(Event); ok {
		r.LogEvent(e.ctx, e.Sender, event)
	}

	r.mux.RLock()
	defer r.mux.RUnlock()
	if r.enabled && e.Receiver != nil {
		r.Send(e)
	}
}

func (r *router) Register(a Addressable) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.contacts[a.Address()] = true
}

func (r *router) Send(e Envelope) {
	if e.Receiver != nil && e.Receiver.IsBroadcast() {
		r.Broadcast(e)
	}
}

func (r *router) Unregister(a Addressable) {
	r.mux.Lock()
	defer r.mux.Unlock()
	delete(r.contacts, a.Address())
}
