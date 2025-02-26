package labor

import "testing"

func TestEventMessage_Event(t *testing.T) {
	tests := []struct {
		name string
		msg  EventMessage
		want Event
	}{
		{
			name: "simple",
			msg:  EventMessage{event: "simple"},
			want: "simple",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := EventMessage{
				event: tt.msg.event,
			}
			if got := e.Event(); got != tt.want {
				t.Errorf("Event() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEventMessage_HasEvent(t *testing.T) {
	tests := []struct {
		name string
		msg  EventMessage
		want bool
	}{
		{
			name: "withSpecificEvent",
			msg:  EventMessage{event: "simple"},
			want: true,
		},
		{
			name: "emptyEvent",
			msg:  EventMessage{},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.msg.HasEvent(); got != tt.want {
				t.Errorf("HasEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
