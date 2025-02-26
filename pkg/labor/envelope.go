package labor

import "context"

type Envelope struct {
	ctx      context.Context
	Sender   string
	Receiver string
	Message  any
}
