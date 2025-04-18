package cedar

import (
	"time"

	"github.com/cedar-policy/cedar-go/internal/mapset"
	"github.com/cedar-policy/cedar-go/types"
	"golang.org/x/exp/constraints"
)

//revive:disable:exported

//  _____
// |_   _|   _ _ __   ___  ___
//   | || | | | '_ \ / _ \/ __|
//   | || |_| | |_) |  __/\__ \
//   |_| \__, | .__/ \___||___/
//       |___/|_|

// Cedar data types

type Boolean = types.Boolean
type Datetime = types.Datetime
type Decimal = types.Decimal
type Duration = types.Duration
type EntityUID = types.EntityUID
type IPAddr = types.IPAddr
type Long = types.Long
type Record = types.Record
type RecordMap = types.RecordMap
type Set = types.Set
type String = types.String

// Other Cedar types

type Entity = types.Entity
type EntityMap = types.EntityMap
type EntityType = types.EntityType
type EntityUIDSet = types.EntityUIDSet
type Pattern = types.Pattern
type Wildcard = types.Wildcard

// cedar-go types

type EntityGetter = types.EntityGetter
type Value = types.Value

type Request = types.Request
type Decision = types.Decision
type Diagnostic = types.Diagnostic
type DiagnosticReason = types.DiagnosticReason
type DiagnosticError = types.DiagnosticError

const (
	Allow = types.Allow
	Deny  = types.Deny
)

type Effect = types.Effect

const (
	Permit = types.Permit
	Forbid = types.Forbid
)

type Annotations = types.Annotations

type Position = types.Position

//   ____                _              _
//  / ___|___  _ __  ___| |_ __ _ _ __ | |_ ___
// | |   / _ \| '_ \/ __| __/ _` | '_ \| __/ __|
// | |__| (_) | | | \__ \ || (_| | | | | |_\__ \
//  \____\___/|_| |_|___/\__\__,_|_| |_|\__|___/

const (
	True  = types.True
	False = types.False
)

//revive:enable:exported

//   ____                _                   _
//  / ___|___  _ __  ___| |_ _ __ _   _  ___| |_ ___  _ __ ___
// | |   / _ \| '_ \/ __| __| '__| | | |/ __| __/ _ \| '__/ __|
// | |__| (_) | | | \__ \ |_| |  | |_| | (__| || (_) | |  \__ \
//  \____\___/|_| |_|___/\__|_|   \__,_|\___|\__\___/|_|  |___/

// NewDatetimeFromMillis returns a Datetime from milliseconds
func NewDatetimeFromMillis(ms int64) Datetime {
	return types.NewDatetimeFromMillis(ms)
}

// NewDurationFromMillis returns a Duration from milliseconds
func NewDurationFromMillis(ms int64) Duration {
	return types.NewDurationFromMillis(ms)
}

// NewDuration returns a Cedar Duration from a Go time.Duration
func NewDuration(d time.Duration) Duration {
	return types.NewDuration(d)
}

// NewDatetime returns a Cedar Datetime from a Go time.Time value
func NewDatetime(t time.Time) Datetime {
	return types.NewDatetime(t)
}

// NewEntityUID returns an EntityUID given an EntityType and identifier
func NewEntityUID(typ EntityType, id String) EntityUID {
	return types.NewEntityUID(typ, id)
}

// NewEntityUIDSet returns an immutable EntityUIDSet ready for use.
func NewEntityUIDSet(args ...EntityUID) EntityUIDSet {
	return mapset.Immutable[EntityUID](args...)
}

// NewPattern permits for the programmatic construction of a Pattern out of a slice of pattern components.
// The pattern components may be one of string, cedar.String, or cedar.Wildcard.  Any other types will
// cause a panic.
func NewPattern(components ...any) Pattern {
	return types.NewPattern(components...)
}

// NewRecord returns an immutable Record given a Go map of Strings to Values
func NewRecord(r RecordMap) Record {
	return types.NewRecord(r)
}

// NewSet returns an immutable Set given a variadic set of Values. Duplicates are removed and order is not preserved.
func NewSet(s ...types.Value) Set {
	return types.NewSet(s...)
}

// NewDecimal returns a Decimal value of i * 10^exponent.
func NewDecimal(i int64, exponent int) (Decimal, error) {
	return types.NewDecimal(i, exponent)
}

// NewDecimalFromInt returns a Decimal with the whole integer value provided
func NewDecimalFromInt[T constraints.Signed](i T) (Decimal, error) {
	return types.NewDecimalFromInt(i)
}

// NewDecimalFromFloat returns a Decimal that approximates the given floating point value.
// The value of the Decimal is calculated by multiplying it by 10^4, truncating it to
// an int64 representation to cut off any digits beyond the four allowed, and passing it
// as an integer to NewDecimal() with -4 as the exponent.
//
// WARNING: decimal representations of more than 6 significant digits for float32s and 15
// significant digits for float64s can be lossy in terms of precision. To create a precise
// Decimal above those sizes, use the NewDecimal constructor.
func NewDecimalFromFloat[T constraints.Float](f T) (Decimal, error) {
	return types.NewDecimalFromFloat(f)
}
