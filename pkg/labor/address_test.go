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
				kind:     "factory",
				id:       "root",
			},
			want: &Address{
				parent:   nil,
				location: LocalLocation,
				kind:     "factory",
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
			name:   "root",
			parent: NewAddress(LocalLocation, "factory", "root"),
			args: args{
				kind: "shed",
				id:   "shed",
			},
			want: NewAddress(LocalLookupAddress, "factory", "root").Child("shed", "shed"),
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
			name:    "factory",
			address: NewAddress(LocalLocation, "factory", "factory"),
			want:    false,
		},
		{
			name:    "shed",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed"),
			want:    true,
		},
		{
			name:    "process",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed").Child("process", "process"),
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
			name:    "factory",
			address: NewAddress(LocalLocation, "factory", "factory"),
			want:    nil,
		},
		{
			name:    "shed",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed"),
			want:    &Address{parent: nil, location: LocalLocation, kind: "factory", id: "factory"},
		},
		{
			name:    "process",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed").Child("process", "process"),
			want:    &Address{parent: &Address{parent: nil, location: LocalLocation, kind: "factory", id: "factory"}, kind: "shed", id: "shed"},
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
			name:    "factory",
			address: NewAddress(LocalLocation, "factory", "factory"),
			want:    "local/factory/factory",
		},
		{
			name:    "shed",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed"),
			want:    "local/factory/factory/shed/shed",
		},
		{
			name:    "process",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed").Child("process", "process"),
			want:    "local/factory/factory/shed/shed/process/process",
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
			name:    "factory",
			address: NewAddress(LocalLocation, "factory", "factory"),
			want:    slog.StringValue("local/factory/factory"),
		},
		{
			name:    "shed",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed"),
			want:    slog.StringValue("local/factory/factory/shed/shed"),
		},
		{
			name:    "process",
			address: NewAddress(LocalLocation, "factory", "factory").Child("shed", "shed").Child("process", "process"),
			want:    slog.StringValue("local/factory/factory/shed/shed/process/process"),
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
