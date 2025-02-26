package labor

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestEvent_LogValue(t *testing.T) {
	type args struct {
		sender string
	}
	tests := []struct {
		name  string
		event Event
		args  args
		want  slog.Attr
	}{
		{
			name:  "message",
			event: Event{Message: "message"},
			args:  args{"sender"},
			want: slog.Group("event",
				slog.String("sender", "sender"),
				slog.String("category", ""),
				slog.String("type", "")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.LogValue(tt.args.sender); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LogValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_String(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{
			name:  "message",
			event: Event{Message: "message"},
			want:  "message",
		},
		{
			name:  "operator",
			event: operatorStartedEvent,
			want:  "operator started",
		},
		{
			name:  "scheduler",
			event: schedulerStartedEvent,
			want:  "scheduler started",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
