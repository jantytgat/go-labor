package labor

import (
	"log/slog"
	"strings"
)

const (
	addressSeparator          = "/"
	LocalAddress              = "local"
	LocalLocation    Location = LocalAddress
)

type Addressable interface {
	Address() *Address
}

type Location string

func (l Location) String() string {
	return string(l)
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

func (a *Address) LogValue() slog.Value {
	return slog.StringValue(a.String())
}
