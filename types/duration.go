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
)

const (
	convertDays    int64 = 86400000
	convertHours   int64 = 3600000
	convertMinutes int64 = 60000
	convertSeconds int64 = 1000
)

var unitToMillis = map[string]int64{
	"d":  convertDays,
	"h":  convertHours,
	"m":  convertMinutes,
	"s":  convertSeconds,
	"ms": 1,
}

var unitOrder = []string{"d", "h", "m", "s", "ms"}

// A Duration is a value representing a span of time in milliseconds.
type Duration struct {
	Value int64
}

// UnsafeDuration creates a duration via unsafe conversion from int, int64, float64
func UnsafeDuration[T int | int64 | float64](v T) Duration {
	return Duration{Value: int64(v)}
}

// FromStdDuration creates a duration from a Go stdlib time.Duration
func FromStdDuration(d time.Duration) Duration {
	return Duration{Value: d.Milliseconds()}
}

// ParseDuration takes a string representation of a duration and parses it into a Duration
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

	return Duration{Value: negative * total}, nil
}

func (a Duration) Equal(bi Value) bool {
	b, ok := bi.(Duration)
	return ok && a == b
}

func (a Duration) Less(bi Value) bool {
	b, ok := bi.(Duration)
	return ok && a.Value < b.Value
}

func (a Duration) LessEqual(bi Value) bool {
	b, ok := bi.(Duration)
	return ok && a.Value <= b.Value
}

// MarshalCedar produces a valid MarshalCedar language representation of the Duration, e.g. `decimal("12.34")`.
func (v Duration) MarshalCedar() []byte { return []byte(`duration("` + v.String() + `")`) }

// String produces a string representation of the Decimal, e.g. `12.34`.
func (v Duration) String() string {
	var res bytes.Buffer
	if v.Value == 0 {
		return "0ms"
	}

	remaining := v.Value
	if v.Value < 0 {
		remaining = -v.Value
		res.WriteByte('-')
	}

	days := remaining / convertDays
	if days > 0 {
		res.WriteString(strconv.FormatInt(days, 10))
		res.WriteByte('d')
	}
	remaining %= convertDays

	hours := remaining / convertHours
	if hours > 0 {
		res.WriteString(strconv.FormatInt(hours, 10))
		res.WriteByte('h')
	}
	remaining %= convertHours

	minutes := remaining / convertMinutes
	if minutes > 0 {
		res.WriteString(strconv.FormatInt(minutes, 10))
		res.WriteByte('m')
	}
	remaining %= convertMinutes

	seconds := remaining / convertSeconds
	if seconds > 0 {
		res.WriteString(strconv.FormatInt(seconds, 10))
		res.WriteByte('s')
	}
	remaining %= convertSeconds

	if remaining > 0 {
		res.WriteString(strconv.FormatInt(remaining, 10))
		res.WriteString("ms")
	}

	return res.String()
}

func (v *Duration) UnmarshalJSON(b []byte) error {
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		// NOTE: cedar supports two other forms, for now we're only supporting the smallest implicit and explicit form.
		// The following are not supported:
		// "duration(\"1d2h3m4s5ms\")"
		// {"fn":"duration","arg":"1d2h3m4s5ms"}
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			return errJSONExtNotFound
		}
		if res.Extn.Fn != "duration" {
			return errJSONExtFnMatch
		}
		arg = res.Extn.Arg
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
func (v Duration) deepClone() Value { return v }

func (v Duration) ToDays() int64 {
	return v.Value / convertDays
}

func (v Duration) ToHours() int64 {
	return v.Value / convertHours
}

func (v Duration) ToMinutes() int64 {
	return v.Value / convertMinutes
}

func (v Duration) ToSeconds() int64 {
	return v.Value / convertSeconds
}

func (v Duration) ToMilliseconds() int64 {
	return v.Value
}
