package types

import (
	"fmt"
)

var ErrDatetime = fmt.Errorf("error parsing datetime value")
var ErrDecimal = fmt.Errorf("error parsing decimal value")
var ErrDuration = fmt.Errorf("error parsing duration value")
var ErrIP = fmt.Errorf("error parsing ip value")

// Value defines the interface for all Cedar values (String, Long, Set, Record, Boolean, etc ...)
type Value interface {
	// String produces a string representation of the Value.
	String() string
	// MarshalCedar produces a valid MarshalCedar language representation of the Value.
	MarshalCedar() []byte
	// ExplicitMarshalJSON marshals the Value into JSON using the explicit (if
	// applicable) JSON form, which is necessary for marshalling values within
	// Sets or Records where the type is not defined.
	ExplicitMarshalJSON() ([]byte, error)
	Equal(Value) bool
	deepClone() Value
}
