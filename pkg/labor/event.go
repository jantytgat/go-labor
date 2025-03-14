package labor

import (
	"log/slog"
)

const (
	laborEventCategory      = "labor"
	managerKind        Kind = "manager"
	operatorKind       Kind = "operators"

	LevelTrace slog.Level = -8
)

var (
	emptySequenceEvent             = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: slog.LevelWarn, Message: "empty job sequence"}
	managerEnabledEvent            = Event{Category: laborEventCategory, Type: managerKind.String(), Level: slog.LevelDebug, Message: "manager enabled"}
	managerDisabledEvent           = Event{Category: laborEventCategory, Type: managerKind.String(), Level: slog.LevelDebug, Message: "manager disabled"}
	managerReceivedJobEvent        = Event{Category: laborEventCategory, Type: managerKind.String(), Level: LevelTrace, Message: "manager received job"}
	processorAvailableEvent        = Event{Category: laborEventCategory, Type: processKind.String(), Level: slog.LevelDebug, Message: "processor available"}
	processorHandleProcessEvent    = Event{Category: laborEventCategory, Type: processKind.String(), Level: LevelTrace, Message: "handle process"}
	operatorAvailableEvent         = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: slog.LevelDebug, Message: "operator available"}
	operatorReceivedJobEvent       = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: slog.LevelDebug, Message: "operator received job"}
	operatorCompletedJobEvent      = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: slog.LevelDebug, Message: "operator completed job"}
	operatorSequenceStartEvent     = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: LevelTrace, Message: "operator started sequence"}
	operatorSequenceCompletedEvent = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: LevelTrace, Message: "operator completed sequence"}
	operatorProcessStartEvent      = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: LevelTrace, Message: "operator started process"}
	operatorProcessCompletedEvent  = Event{Category: laborEventCategory, Type: operatorKind.String(), Level: LevelTrace, Message: "operator completed process"}
	registeredEvent                = Event{Category: laborEventCategory, Type: managerKind.String(), Level: slog.LevelDebug, Message: "registered address"}
	unsupportedMessageEvent        = Event{Category: laborEventCategory, Type: managerKind.String(), Level: slog.LevelWarn, Message: "unsupported message"}
)

type Kind string

func (k Kind) String() string {
	return string(k)
}

type Event struct {
	Category string
	Type     string
	Message  string
	Info     any
	Level    slog.Level
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
