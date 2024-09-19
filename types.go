package cedar

import (
	"time"

	"github.com/cedar-policy/cedar-go/types"
)

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

type Entities = types.Entities
type Entity = types.Entity
type EntityType = types.EntityType
type Pattern = types.Pattern
type Wildcard = types.Wildcard

// cedar-go types

type Value = types.Value

//   ____                _              _
//  / ___|___  _ __  ___| |_ __ _ _ __ | |_ ___
// | |   / _ \| '_ \/ __| __/ _` | '_ \| __/ __|
// | |__| (_) | | | \__ \ || (_| | | | | |_\__ \
//  \____\___/|_| |_|___/\__\__,_|_| |_|\__|___/

const (
	True  = types.True
	False = types.False
)

const (
	DecimalPrecision = types.DecimalPrecision
)

//   ____                _                   _
//  / ___|___  _ __  ___| |_ _ __ _   _  ___| |_ ___  _ __ ___
// | |   / _ \| '_ \/ __| __| '__| | | |/ __| __/ _ \| '__/ __|
// | |__| (_) | | | \__ \ |_| |  | |_| | (__| || (_) | |  \__ \
//  \____\___/|_| |_|___/\__|_|   \__,_|\___|\__\___/|_|  |___/

// DatetimeFromMillis returns a Datetime from milliseconds
func DatetimeFromMillis(ms int64) Datetime {
	return types.DatetimeFromMillis(ms)
}

// DurationFromMillis returns a Duration from milliseconds
func DurationFromMillis(ms int64) Duration {
	return types.DurationFromMillis(ms)
}

// FromStdDuration returns a Cedar Duration from a Go time.Duration
func FromStdDuration(d time.Duration) Duration {
	return types.FromStdDuration(d)
}

// FromStdTime returns a Cedar Datetime from a Go time.Time value
func FromStdTime(t time.Time) Datetime {
	return types.FromStdTime(t)
}

// NewEntityUID returns an EntityUID given an EntityType and identifier
func NewEntityUID(typ EntityType, id String) EntityUID {
	return types.NewEntityUID(typ, id)
}

// NewPattern permits for the programmatic construction of a Pattern out of a slice of pattern components.
// The pattern components may be one of string, cedar.String, or cedar.Wildcard.  Any other types will
// cause a panic.
func NewPattern(components ...any) Pattern {
	return types.NewPattern(components)
}

// NewRecord returns an immutable Record given a Go map of Strings to Values
func NewRecord(r RecordMap) Record {
	return types.NewRecord(r)
}

// NewSet returns an immutable Set given a Go slice of Values. Duplicates are removed and order is not preserved.
func NewSet(s []types.Value) Set {
	return types.NewSet(s)
}

// UnsafeDecimal creates a decimal via unsafe conversion from int, int64, float64.
// Precision may be lost and overflows may occur.
func UnsafeDecimal[T int | int64 | float64](v T) Decimal {
	return types.UnsafeDecimal(v)
}
