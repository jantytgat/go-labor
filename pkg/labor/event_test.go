package labor

import "testing"

func TestEvent_String(t *testing.T) {
	tests := []struct {
		name string
		e    Event
		want string
	}{
		{
			name: "simple",
			e:    Event("simple"),
			want: "simple",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
