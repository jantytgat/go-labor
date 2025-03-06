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
		config:             config,
		broadcastListeners: make(map[Addressable]bool),
	}
}

type router struct {
	config             routerConfig
	broadcastListeners map[Addressable]bool
	eventLogger        *slog.Logger
	mux                sync.RWMutex
}

func (r *router) broadcast(e envelope) {
	r.mux.RLock()
	defer r.mux.RUnlock()
	for contact, broadcast := range r.broadcastListeners {
		if !broadcast {
			continue
		}
		e.Receiver = contact
		r.forward(e)
	}
}

func (r *router) logEvent(ctx context.Context, sender Addressable, event Event) {
	r.config.EventLogger.LogAttrs(
		ctx,
		r.config.EventLogLevel,
		event.String(),
		event.LogValue(sender.Address()))
}

func (r *router) Send(e envelope) {
	if event, ok := e.Message.(Event); ok {
		r.logEvent(e.ctx, e.Sender, event)
	}

	r.send(e)
}

func (r *router) Register(a Addressable) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.broadcastListeners[a] = true
}

func (r *router) send(e envelope) {
	if e.Receiver == nil {
		return
	}

	switch e.Receiver.Address().IsBroadcast() {
	case true:
		go r.broadcast(e)
	case false:
		r.forward(e)
	}
}

func (r *router) forward(e envelope) {
	e.Receiver.Receive(e)
}

func (r *router) Unregister(a Addressable) {
	r.mux.Lock()
	defer r.mux.Unlock()
	delete(r.broadcastListeners, a)
}
