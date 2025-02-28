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
		res: &responseHandler{
			address: NewAddress(LocalLocation, customerKind, name),
			output:  make(chan Envelope),
		},
	}
}

type Customer struct {
	res *responseHandler
}

func (c *Customer) Send(ctx context.Context, job Request, m *Manager) error {
	if c.res == nil {
		return fmt.Errorf("customer not properly initialized")
	}

	m.router.Send(Envelope{
		ctx:      ctx,
		Sender:   c.res,
		Receiver: m.scheduler,
		Message:  job,
	})
	return nil
}

func (c *Customer) Receive(ctx context.Context) any {
	select {
	case <-ctx.Done():
		return nil
	case e := <-c.res.output:
		return e.Message
	}
}

type responseHandler struct {
	address *Address
	output  chan Envelope
}

func (r *responseHandler) Address() *Address {
	return r.address
}

func (r *responseHandler) Receive(e Envelope) {
	r.output <- e
}
