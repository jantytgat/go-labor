package labor

import "context"

type Envelope struct {
	ctx      context.Context
	Sender   *Address
	Receiver *Address
	Message  any
}
