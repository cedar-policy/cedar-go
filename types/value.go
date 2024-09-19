package types

import (
	"fmt"
)

var ErrDatetime = fmt.Errorf("error parsing datetime value")
var ErrDecimal = fmt.Errorf("error parsing decimal value")
var ErrDuration = fmt.Errorf("error parsing duration value")
var ErrIP = fmt.Errorf("error parsing ip value")
var ErrNotComparable = fmt.Errorf("incompatible types in comparison")

// Value defines the interface for all Cedar values (String, Long, Set, Record, Boolean, etc ...)
//
// Implementations of Value _must_ be able to be safely copied shallowly, which means they must either be immutable
// or be made up of data structures that are free of pointers (e.g. slices and maps).
type Value interface {
	fmt.Stringer
	// MarshalCedar produces a valid MarshalCedar language representation of the Value.
	MarshalCedar() []byte
	// ExplicitMarshalJSON marshals the Value into JSON using the explicit (if
	// applicable) JSON form, which is necessary for marshalling values within
	// Sets or Records where the type is not defined.
	ExplicitMarshalJSON() ([]byte, error)
	Equal(Value) bool
	hash() uint64
}
