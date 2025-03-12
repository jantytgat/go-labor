package labor

import "context"

type Process struct {
	address       *Address
	Name          string
	maxConcurrent int
	Execute       func(ctx context.Context, data any) error
}

func (p *Process) Address() *Address {
	return p.address
}

func (p *Process) Receive(e envelope) {

}
