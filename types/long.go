package types

import (
	"encoding/json"
	"fmt"
)

// A Long is a whole number without decimals that can range from -9223372036854775808 to 9223372036854775807.
type Long int64

func (a Long) Equal(bi Value) bool {
	b, ok := bi.(Long)
	return ok && a == b
}

func (a Long) Less(bi Value) bool {
	b, ok := bi.(Long)
	return ok && a < b
}

func (a Long) LessEqual(bi Value) bool {
	b, ok := bi.(Long)
	return ok && a <= b
}

// ExplicitMarshalJSON marshals the Long into JSON.
func (v Long) ExplicitMarshalJSON() ([]byte, error) { return json.Marshal(v) }

// String produces a string representation of the Long, e.g. `42`.
func (v Long) String() string { return fmt.Sprint(int64(v)) }

// MarshalCedar produces a valid MarshalCedar language representation of the Long, e.g. `42`.
func (v Long) MarshalCedar() []byte {
	return []byte(v.String())
}
func (v Long) deepClone() Value { return v }
