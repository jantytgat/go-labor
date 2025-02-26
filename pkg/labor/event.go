package labor

import "log/slog"

const (
	laborEventCategory = "labor"
)

type Event struct {
	Category string
	Type     string
	Message  string
}

func (e Event) LogValue(sender string) slog.Attr {
	return slog.Group(
		"event",
		slog.String("sender", sender),
		slog.String("category", e.Category),
		slog.String("type", e.Type),
	)
}

func (e Event) String() string {
	return e.Message
}
