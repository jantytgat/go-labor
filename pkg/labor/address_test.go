package labor

import (
	"log/slog"
	"reflect"
	"testing"
)

func TestLocation_String(t *testing.T) {
	tests := []struct {
		name string
		l    Location
		want string
	}{
		{
			name: "local",
			l:    LocalLocation,
			want: LocalAddress,
		},
		{
			name: "broadcast",
			l:    BroadcastLocation,
			want: BroadcastAddress,
		},
		{
			name: "remote",
			l:    "remote",
			want: "remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAddress(t *testing.T) {
	type args struct {
		location Location
		kind     Kind
		id       string
	}
	tests := []struct {
		name string
		args args
		want *Address
	}{
		{
			name: "root",
			args: args{
				location: LocalLocation,
				kind:     "manager",
				id:       "root",
			},
			want: &Address{
				parent:   nil,
				location: LocalLocation,
				kind:     "manager",
				id:       "root",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAddress(tt.args.location, tt.args.kind, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Child(t *testing.T) {
	type args struct {
		kind Kind
		id   string
	}
	tests := []struct {
		name   string
		parent *Address
		args   args
		want   *Address
	}{
		{
			name:   "manager",
			parent: NewAddress(LocalLocation, "manager", "root"),
			args: args{
				kind: "router",
				id:   "root",
			},
			want: NewAddress(LocalAddress, "manager", "root").Child("router", "root"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.parent.Child(tt.args.kind, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Child() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_HasParent(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    bool
	}{
		{
			name:    "manager",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    false,
		},
		{
			name:    "operator",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "true"),
			want:    true,
		},
		{
			name:    "worker",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.HasParent(); got != tt.want {
				t.Errorf("HasParent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Parent(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    *Address
	}{
		{
			name:    "manager",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    nil,
		},
		{
			name:    "operator",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "true"),
			want:    &Address{parent: nil, location: LocalLocation, kind: "manager", id: "root"},
		},
		{
			name:    "worker",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    &Address{parent: &Address{parent: nil, location: LocalLocation, kind: "manager", id: "root"}, kind: "operator", id: "root"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.Parent(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_String(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    string
	}{
		{
			name:    "manager",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    "local/manager/root",
		},
		{
			name:    "router",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "root"),
			want:    "local/manager/root/router/root",
		},
		{
			name:    "worker",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    "local/manager/root/operator/root/worker/1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_LogValue(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    slog.Value
	}{
		{
			name:    "manager",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    slog.StringValue("local/manager/root"),
		},
		{
			name:    "router",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "root"),
			want:    slog.StringValue("local/manager/root/router/root"),
		},
		{
			name:    "worker",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    slog.StringValue("local/manager/root/operator/root/worker/1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.LogValue(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LogValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_IsBroadcast(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    bool
	}{
		{
			name:    "rootNonBroadcast",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    false,
		},
		{
			name:    "rootBroadcast",
			address: NewAddress(BroadcastLocation, "manager", "root"),
			want:    true,
		},
		{
			name:    "nestedNonBroadcast",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "root"),
			want:    false,
		},
		{
			name:    "nestedBroadcast",
			address: NewAddress(BroadcastLocation, "manager", "root").Child("router", "root"),
			want:    true,
		},
		{
			name:    "multiNestedNonBroadcast",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    false,
		},
		{
			name:    "multiNestedBroadcast",
			address: NewAddress(BroadcastLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.IsBroadcast(); got != tt.want {
				t.Errorf("IsBroadcast() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_IsLocal(t *testing.T) {
	tests := []struct {
		name    string
		address *Address
		want    bool
	}{
		{
			name:    "rootLocal",
			address: NewAddress(LocalLocation, "manager", "root"),
			want:    true,
		},
		{
			name:    "rootNonLocal",
			address: NewAddress(BroadcastLocation, "manager", "root"),
			want:    false,
		},
		{
			name:    "nestedLocal",
			address: NewAddress(LocalLocation, "manager", "root").Child("router", "root"),
			want:    true,
		},
		{
			name:    "nestedNonLocal",
			address: NewAddress(BroadcastLocation, "manager", "root").Child("router", "root"),
			want:    false,
		},
		{
			name:    "multiNestedLocal",
			address: NewAddress(LocalLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    true,
		},
		{
			name:    "multiNestedNonLocal",
			address: NewAddress(BroadcastLocation, "manager", "root").Child("operator", "root").Child("worker", "1"),
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.address.IsLocal(); got != tt.want {
				t.Errorf("IsLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}
