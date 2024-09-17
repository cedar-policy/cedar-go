package types

import (
	"encoding/json"
)

// A Boolean is a value that is either true or false.
type Boolean bool

const (
	True  = Boolean(true)
	False = Boolean(false)
)

func (a Boolean) Equal(bi Value) bool {
	b, ok := bi.(Boolean)
	return ok && a == b
}

// String produces a string representation of the Boolean, e.g. `true`.
func (v Boolean) String() string { return string(v.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Boolean, e.g. `true`.
func (v Boolean) MarshalCedar() []byte {
	if v {
		return []byte("true")
	}
	return []byte("false")
}

// ExplicitMarshalJSON marshals the Boolean into JSON.
func (v Boolean) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }

func (v Boolean) hash() uint64 {
	if v {
		return 1
	}
	return 0
}
