package labor

import (
	"context"
	"fmt"
	"sync"
)

const (
	customerKind Kind = "customer"
)

func NewCustomer(name string) *Customer {
	return &Customer{
		Name: name,
		res: &responseHandler{
			address: NewAddress(LocalLocation, customerKind, name),
			output:  make(chan envelope),
		},
		Requests:  0,
		Responses: 0,
	}
}

type Customer struct {
	Name      string
	Requests  int
	Responses int
	res       *responseHandler
	mux       sync.Mutex
}

func (c *Customer) Send(ctx context.Context, job Job, m *Manager) error {
	if c.res == nil {
		return fmt.Errorf("customer not properly initialized")
	}

	if !m.IsEnabled() {
		return fmt.Errorf("manager is not accepting new messages")
	}

	go m.Receive(envelope{
		ctx:     ctx,
		Sender:  c.res,
		Message: job,
	})

	c.mux.Lock()
	c.Requests++
	c.mux.Unlock()
	return nil
}

func (c *Customer) Receive(ctx context.Context) any {
	select {
	case <-ctx.Done():
		return nil
	case e := <-c.res.output:
		c.mux.Lock()
		c.Responses++
		c.mux.Unlock()
		return e.Message
	}
}

func (c *Customer) RequestsTotal() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.Requests
}

func (c *Customer) ResponsesTotal() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.Responses
}

type responseHandler struct {
	address *Address
	output  chan envelope
}

func (r *responseHandler) Address() *Address {
	return r.address
}

func (r *responseHandler) Receive(e envelope) {
	r.output <- e
}
