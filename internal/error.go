package internal

import "fmt"

// These errors are declared here in order to allow the tests outside of the
// types package to assert on the error type returned. One day, we could
// consider making them public.

var ErrDatetime = fmt.Errorf("error parsing datetime value")
var ErrDecimal = fmt.Errorf("error parsing decimal value")
var ErrDuration = fmt.Errorf("error parsing duration value")
var ErrIP = fmt.Errorf("error parsing ip value")
var ErrNotComparable = fmt.Errorf("incompatible types in comparison")
var ErrDurationRange = fmt.Errorf("duration out of range")
