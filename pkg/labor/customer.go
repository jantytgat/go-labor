package labor

import (
	"context"
	"fmt"
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
	}
}

type Customer struct {
	Name      string
	Requests  int
	Responses int
	res       *responseHandler
}

func (c *Customer) Send(ctx context.Context, job Job, m *Manager) error {
	if c.res == nil {
		return fmt.Errorf("customer not properly initialized")
	}

	if !m.IsEnabled() {
		return fmt.Errorf("manager is not accepting new messages")
	}

	m.Receive(envelope{
		ctx:     ctx,
		Sender:  c.res,
		Message: job,
	})
	c.Requests++
	return nil
}

func (c *Customer) Receive(ctx context.Context) any {
	select {
	case <-ctx.Done():
		return nil
	case e := <-c.res.output:
		c.Responses++
		return e.Message
	}
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
