package types

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/internal"
)

// A Long is a whole number without decimals that can range from -9223372036854775808 to 9223372036854775807.
type Long int64

func (l Long) Equal(bi Value) bool {
	b, ok := bi.(Long)
	return ok && l == b
}

func (l Long) LessThan(bi Value) (bool, error) {
	b, ok := bi.(Long)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return l < b, nil
}

func (l Long) LessThanOrEqual(bi Value) (bool, error) {
	b, ok := bi.(Long)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return l <= b, nil
}

// String produces a string representation of the Long, e.g. `42`.
func (l Long) String() string { return fmt.Sprint(int64(l)) }

// MarshalCedar produces a valid MarshalCedar language representation of the Long, e.g. `42`.
func (l Long) MarshalCedar() []byte {
	return []byte(l.String())
}

func (l Long) hash() uint64 {
	return uint64(l)
}
