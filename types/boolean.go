package types

import (
	"encoding/json"
	"fmt"
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
func (v Boolean) TypeName() string { return "bool" }

// String produces a string representation of the Boolean, e.g. `true`.
func (v Boolean) String() string { return v.Cedar() }

// Cedar produces a valid Cedar language representation of the Boolean, e.g. `true`.
func (v Boolean) Cedar() string {
	return fmt.Sprint(bool(v))
}

// ExplicitMarshalJSON marshals the Boolean into JSON.
func (v Boolean) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }
func (v Boolean) deepClone() Value                     { return v }
