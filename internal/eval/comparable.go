package eval

import "github.com/cedar-policy/cedar-go/types"

// ComparableValue provides the interface that must be implemented to
// support operator overloading of <, <=, >, and >=
type ComparableValue interface {
	types.Value

	// LessThan returns true if the lhs is less than the rhs, and an
	// error if the rhs is not comparable to the lhs
	LessThan(types.Value) (bool, error)

	// LessThan returns true if the lhs is less than or equal to the
	// rhs, and an error if the rhs is not comparable to the lhs
	LessThanOrEqual(types.Value) (bool, error)
}
