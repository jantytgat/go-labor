package labor

import (
	"log/slog"
	"strings"
)

const (
	addressSeparator = "/"

	LocalAddress     = "local"
	BroadcastAddress = "broadcast"

	BroadcastLocation Location = Location(BroadcastAddress)
	LocalLocation              = Location(LocalAddress)
)

type Addressable interface {
	Address() *Address
	Receive(Envelope)
}

func NewAddress(location Location, kind Kind, id string) *Address {
	return &Address{
		location: location,
		kind:     kind,
		id:       id,
	}
}

// Address represents the logical id of a component (factory, shed, process, handler) and is used for messaging.
type Address struct {
	parent   *Address
	location Location
	kind     Kind
	id       string
}

func (a *Address) Child(kind Kind, id string) *Address {
	return &Address{
		parent: a,
		kind:   kind,
		id:     id,
	}
}

func (a *Address) HasParent() bool {
	return a.parent != nil
}

func (a *Address) IsBroadcast() bool {
	if a.parent != nil {
		return a.parent.IsBroadcast()
	}
	return a.location == BroadcastLocation
}

func (a *Address) IsLocal() bool {
	if a.parent != nil {
		return a.parent.IsLocal()
	}
	return a.location == LocalLocation
}

func (a *Address) LogValue() slog.Value {
	return slog.StringValue(a.String())
}

func (a *Address) Parent() *Address {
	return a.parent
}

func (a *Address) String() string {
	if a.HasParent() {
		return strings.Join([]string{a.parent.String(), a.kind.String(), a.id}, addressSeparator)
	} else {
		return strings.Join([]string{a.location.String(), a.kind.String(), a.id}, addressSeparator)
	}
}
