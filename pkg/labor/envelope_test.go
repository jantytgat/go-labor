package labor

import "testing"

type hasEventMessage struct{}

func (m hasEventMessage) HasEvent() bool {
	return true
}
func (m hasEventMessage) Event() Event {
	return "event"
}

func TestEnvelope_Event(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope
		want     Event
	}{
		{
			name: "hasEvent",
			envelope: Envelope{
				Message: hasEventMessage{},
			},
			want: "event",
		},
		{
			name: "noEvent",
			envelope: Envelope{
				Sender:   "",
				Receiver: "",
				Message:  nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Envelope{
				Sender:   tt.envelope.Sender,
				Receiver: tt.envelope.Receiver,
				Message:  tt.envelope.Message,
			}
			if got := e.Event(); got != tt.want {
				t.Errorf("Event() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvelope_HasEvent(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope
		want     bool
	}{
		{
			name: "hasEvent",
			envelope: Envelope{
				Message: hasEventMessage{},
			},
			want: true,
		},
		{
			name: "noEvent",
			envelope: Envelope{
				Sender:   "",
				Receiver: "",
				Message:  nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Envelope{
				Sender:   tt.envelope.Sender,
				Receiver: tt.envelope.Receiver,
				Message:  tt.envelope.Message,
			}
			if got := e.HasEvent(); got != tt.want {
				t.Errorf("HasEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
