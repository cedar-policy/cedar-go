package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
	"unicode"

	"github.com/cedar-policy/cedar-go/internal/consts"
)

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

// FromStdDuration returns a Cedar Duration from a Go time.Duration
func FromStdDuration(d time.Duration) Duration {
	return Duration{value: d.Milliseconds()}
}

// DurationFromMillis returns a Duration from milliseconds
func DurationFromMillis(ms int64) Duration {
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
		return Duration{}, fmt.Errorf("%w: string too short", ErrDuration)
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
				return Duration{}, fmt.Errorf("%w: overflow", ErrDuration)
			}
			hasValue = true
			i++
		} else if s[i] == 'd' || s[i] == 'h' || s[i] == 'm' || s[i] == 's' {
			if !hasValue {
				return Duration{}, fmt.Errorf("%w: unit found without quantity", ErrDuration)
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
				return Duration{}, fmt.Errorf("%w: unexpected unit '%s'", ErrDuration, unit)
			}

			total = total + value*unitToMillis[unit]
			i++
			hasValue = false
			value = 0
		} else {
			return Duration{}, fmt.Errorf("%w: unexpected character %s", ErrDuration, strconv.QuoteRune(rune(s[i])))
		}
	}

	// We didn't have a trailing unit
	if hasValue {
		return Duration{}, fmt.Errorf("%w: expected unit", ErrDuration)
	}

	// We still have characters left, but no more units to assign.
	if i < len(s) {
		return Duration{}, fmt.Errorf("%w: invalid duration", ErrDuration)
	}

	return Duration{value: negative * total}, nil
}

// Equal returns true if the input represents the same duration
func (a Duration) Equal(bi Value) bool {
	b, ok := bi.(Duration)
	return ok && a == b
}

// LessThan returns true if value is less than the argument and they
// are both Duration values, or an error indicating they aren't
// comparable otherwise
func (a Duration) LessThan(bi Value) (bool, error) {
	b, ok := bi.(Duration)
	if !ok {
		return false, ErrNotComparable
	}
	return a.value < b.value, nil
}

// LessThan returns true if value is less than or equal to the
// argument and they are both Duration values, or an error indicating
// they aren't comparable otherwise
func (a Duration) LessThanOrEqual(bi Value) (bool, error) {
	b, ok := bi.(Duration)
	if !ok {
		return false, ErrNotComparable
	}
	return a.value <= b.value, nil
}

// MarshalCedar produces a valid MarshalCedar language representation of the Duration, e.g. `decimal("12.34")`.
func (v Duration) MarshalCedar() []byte { return []byte(`duration("` + v.String() + `")`) }

// String produces a string representation of the Duration
func (v Duration) String() string {
	var res bytes.Buffer
	if v.value == 0 {
		return "0ms"
	}

	remaining := v.value
	if v.value < 0 {
		remaining = -v.value
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
// It is capable of unmarshaling 4 different representations supported by Cedar
// - { "__extn": { "fn": "duration", "arg": "1h10m" }}
// - { "fn": "duration", "arg": "1h10m" }
// - "duration(\"1h10m\")"
// - "1h10m"
func (v *Duration) UnmarshalJSON(b []byte) error {
	var arg string
	if bytes.HasPrefix(b, []byte(`"duration(\"`)) && bytes.HasSuffix(b, []byte(`\")"`)) {
		arg = string(b[12 : len(b)-4])
	} else if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			// If we didn't find an Extn, maybe it's just an extn.
			var res2 extn
			json.Unmarshal(b, &res2)
			// We've tried Ext.Fn and Fn, so no good.
			if res2.Fn == "" {
				return errJSONExtNotFound
			}
			if res2.Fn != "duration" {
				return errJSONExtFnMatch
			}
			arg = res2.Arg
		} else if res.Extn.Fn != "duration" {
			return errJSONExtFnMatch
		} else {
			arg = res.Extn.Arg
		}
	}
	vv, err := ParseDuration(arg)
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

// MarshalJSON marshals the Duration into JSON using the implicit form.
func (v Duration) MarshalJSON() ([]byte, error) { return []byte(`"` + v.String() + `"`), nil }

// ExplicitMarshalJSON marshals the Decimal into JSON using the explicit form.
func (v Duration) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "duration",
			Arg: v.String(),
		},
	})
}

// ToDays returns the number of days this Duration represents,
// truncating when fractional
func (v Duration) ToDays() int64 {
	return v.value / consts.MillisPerDay
}

// ToHours returns the number of hours this Duration represents,
// truncating when fractional
func (v Duration) ToHours() int64 {
	return v.value / consts.MillisPerHour
}

// ToMinutes returns the number of minutes this Duration represents,
// truncating when fractional
func (v Duration) ToMinutes() int64 {
	return v.value / consts.MillisPerMinute
}

// ToSeconds returns the number of seconds this Duration represents,
// truncating when fractional
func (v Duration) ToSeconds() int64 {
	return v.value / consts.MillisPerSecond
}

// ToMilliseconds returns the number of milliseconds this Duration
// represents
func (v Duration) ToMilliseconds() int64 {
	return v.value
}

func (v Duration) hash() uint64 {
	return uint64(v.value)
}
