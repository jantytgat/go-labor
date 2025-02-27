package labor

import "context"

type Envelope struct {
	ctx      context.Context
	Sender   Addressable
	Receiver Addressable
	Message  any
}
