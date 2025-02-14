package types

import (
	"fmt"
)

// Value defines the interface for all Cedar values (String, Long, Set, Record, Boolean, etc ...)
//
// Implementations of Value _must_ be able to be safely copied shallowly, which means they must either be immutable
// or be made up of data structures that are free of pointers (e.g. slices and maps).
type Value interface {
	fmt.Stringer
	// MarshalCedar produces a valid MarshalCedar language representation of the Value.
	MarshalCedar() []byte
	Equal(Value) bool
	hash() uint64
}
