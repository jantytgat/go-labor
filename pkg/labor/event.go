package labor

import (
	"log/slog"
)

const (
	laborEventCategory = "labor"
)

type Event struct {
	Category string
	Type     string
	Message  string
	Info     any
}

func (e Event) LogValue(sender *Address) slog.Attr {
	return slog.Group(
		"event",
		slog.String("source", sender.String()),
		slog.String("category", e.Category),
		slog.String("type", e.Type),
		slog.Any("info", e.Info),
	)
}

func (e Event) String() string {
	return e.Message
}

func (e Event) WithInfo(info any) Event {
	e.Info = info
	return e
}
