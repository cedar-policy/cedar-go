package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode"

	"github.com/cedar-policy/cedar-go/internal"
	"github.com/cedar-policy/cedar-go/internal/consts"
)

var errDuration = internal.ErrDuration

var unitToMillis = map[string]int64{
	"d":  consts.MillisPerDay,
	"h":  consts.MillisPerHour,
	"m":  consts.MillisPerMinute,
	"s":  consts.MillisPerSecond,
	"ms": 1,
}

var unitOrder = []string{"d", "h", "m", "s", "ms"}

// A Duration is a value representing a span of time in milliseconds.
type Duration struct {
	value int64
}

// NewDuration returns a Cedar Duration from a Go time.Duration
func NewDuration(d time.Duration) Duration {
	return Duration{value: d.Milliseconds()}
}

// NewDurationFromMillis returns a Duration from milliseconds
func NewDurationFromMillis(ms int64) Duration {
	return Duration{value: ms}
}

// ParseDuration parses a Cedar Duration from a string
//
// Cedar RFC 80 defines a valid duration string as collapsed sequence
// of quantity-unit pairs, possibly with a leading `-`, indicating a
// negative duration.
// The units must appear in order from longest timeframe to smallest.
// - d: days
// - h: hours
// - m: minutes
// - s: seconds
// - ms: milliseconds
func ParseDuration(s string) (Duration, error) {
	// Check for empty string.
	if len(s) <= 1 {
		return Duration{}, fmt.Errorf("%w: string too short", errDuration)
	}

	i := 0
	unitI := 0

	negative := int64(1)
	if s[i] == '-' {
		negative = int64(-1)
		i++
	}

	var (
		total    int64
		value    int64
		unit     string
		hasValue bool
	)

	// ([0-9]+)(d|h|m|s|ms) ...
	for i < len(s) && unitI < len(unitOrder) {
		if unicode.IsDigit(rune(s[i])) {
			value = value*10 + int64(s[i]-'0')

			// check overflow
			if value > math.MaxInt32 {
				return Duration{}, fmt.Errorf("%w: overflow", errDuration)
			}
			hasValue = true
			i++
		} else if s[i] == 'd' || s[i] == 'h' || s[i] == 'm' || s[i] == 's' {
			if !hasValue {
				return Duration{}, fmt.Errorf("%w: unit found without quantity", errDuration)
			}

			// is it ms?
			if s[i] == 'm' && i+1 < len(s) && s[i+1] == 's' {
				unit = "ms"
				i++
			} else {
				unit = s[i : i+1]
			}

			unitOK := false
			for !unitOK && unitI < len(unitOrder) {
				if unit == unitOrder[unitI] {
					unitOK = true
				}
				unitI++
			}

			if !unitOK {
				return Duration{}, fmt.Errorf("%w: unexpected unit '%s'", errDuration, unit)
			}

			total = total + value*unitToMillis[unit]
			i++
			hasValue = false
			value = 0
		} else {
			return Duration{}, fmt.Errorf("%w: unexpected character %s", errDuration, strconv.QuoteRune(rune(s[i])))
		}
	}

	// We didn't have a trailing unit
	if hasValue {
		return Duration{}, fmt.Errorf("%w: expected unit", errDuration)
	}

	// We still have characters left, but no more units to assign.
	if i < len(s) {
		return Duration{}, fmt.Errorf("%w: invalid duration", errDuration)
	}

	return Duration{value: negative * total}, nil
}

// Equal returns true if the input represents the same duration
func (d Duration) Equal(bi Value) bool {
	b, ok := bi.(Duration)
	return ok && d == b
}

// LessThan returns true if value is less than the argument and they
// are both Duration values, or an error indicating they aren't
// comparable otherwise
func (d Duration) LessThan(bi Value) (bool, error) {
	b, ok := bi.(Duration)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return d.value < b.value, nil
}

// LessThan returns true if value is less than or equal to the
// argument and they are both Duration values, or an error indicating
// they aren't comparable otherwise
func (d Duration) LessThanOrEqual(bi Value) (bool, error) {
	b, ok := bi.(Duration)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return d.value <= b.value, nil
}

// MarshalCedar produces a valid MarshalCedar language representation of the Duration, e.g. `decimal("12.34")`.
func (d Duration) MarshalCedar() []byte { return []byte(`duration("` + d.String() + `")`) }

// String produces a string representation of the Duration
func (d Duration) String() string {
	var res bytes.Buffer
	if d.value == 0 {
		return "0ms"
	}

	remaining := d.value
	if d.value < 0 {
		remaining = -d.value
		res.WriteByte('-')
	}

	days := remaining / consts.MillisPerDay
	if days > 0 {
		res.WriteString(strconv.FormatInt(days, 10))
		res.WriteByte('d')
	}
	remaining %= consts.MillisPerDay

	hours := remaining / consts.MillisPerHour
	if hours > 0 {
		res.WriteString(strconv.FormatInt(hours, 10))
		res.WriteByte('h')
	}
	remaining %= consts.MillisPerHour

	minutes := remaining / consts.MillisPerMinute
	if minutes > 0 {
		res.WriteString(strconv.FormatInt(minutes, 10))
		res.WriteByte('m')
	}
	remaining %= consts.MillisPerMinute

	seconds := remaining / consts.MillisPerSecond
	if seconds > 0 {
		res.WriteString(strconv.FormatInt(seconds, 10))
		res.WriteByte('s')
	}
	remaining %= consts.MillisPerSecond

	if remaining > 0 {
		res.WriteString(strconv.FormatInt(remaining, 10))
		res.WriteString("ms")
	}

	return res.String()
}

// UnmarshalJSON implements encoding/json.Unmarshaler for Duration
//
// It is capable of unmarshaling 3 different representations supported by Cedar
//   - { "__extn": { "fn": "duration", "arg": "1h10m" }}
//   - { "fn": "duration", "arg": "1h10m" }
//   - "1h10m"
func (d *Duration) UnmarshalJSON(b []byte) error {
	vv, err := unmarshalExtensionValue(b, "duration", ParseDuration)
	if err != nil {
		return err
	}

	*d = vv
	return nil
}

// MarshalJSON marshals the Duration into JSON using the explicit form.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "duration",
			Arg: d.String(),
		},
	})
}

// ToDays returns the number of days this Duration represents,
// truncating when fractional
func (d Duration) ToDays() int64 {
	return d.value / consts.MillisPerDay
}

// ToHours returns the number of hours this Duration represents,
// truncating when fractional
func (d Duration) ToHours() int64 {
	return d.value / consts.MillisPerHour
}

// ToMinutes returns the number of minutes this Duration represents,
// truncating when fractional
func (d Duration) ToMinutes() int64 {
	return d.value / consts.MillisPerMinute
}

// ToSeconds returns the number of seconds this Duration represents,
// truncating when fractional
func (d Duration) ToSeconds() int64 {
	return d.value / consts.MillisPerSecond
}

// ToMilliseconds returns the number of milliseconds this Duration
// represents
func (d Duration) ToMilliseconds() int64 {
	return d.value
}

// Duration returns a time.Duration representation of a Duration.  An error
// is returned if the duration cannot be converted to a time.Duration.
func (d Duration) Duration() (time.Duration, error) {
	if d.value > math.MaxInt64/1000 {
		return 0, internal.ErrDurationRange
	}
	if d.value < math.MinInt64/1000 {
		return 0, internal.ErrDurationRange
	}
	return time.Millisecond * time.Duration(d.value), nil
}

func (d Duration) hash() uint64 {
	return uint64(d.value)
}
