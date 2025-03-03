package labor

import "context"

type envelope struct {
	ctx      context.Context
	Sender   Addressable
	Receiver Addressable
	Message  any
}
