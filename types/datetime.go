package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"unicode"

	"github.com/cedar-policy/cedar-go/internal"
)

var errDatetime = internal.ErrDatetime

// Datetime represents a Cedar datetime value
type Datetime struct {
	// value is a timestamp in milliseconds
	value int64
}

// NewDatetime returns a Cedar Datetime from a Go time.Time value
func NewDatetime(t time.Time) Datetime {
	return Datetime{value: t.UnixMilli()}
}

// NewDatetimeFromMillis returns a Datetime from a count of milliseconds since
// January 1, 1970 @ 00:00:00 UTC.
func NewDatetimeFromMillis(ms int64) Datetime {
	return Datetime{value: ms}
}

// ParseDatetime returns a Cedar datetime when the argument provided
// represents a compatible datetime or an error
//
// Cedar RFC 80 defines valid datetime strings as one of:
//
// - "YYYY-MM-DD" (date only, with implied time 00:00:00, UTC)
// - "YYYY-MM-DDThh:mm:ssZ" (date and time, UTC)
// - "YYYY-MM-DDThh:mm:ss.SSSZ" (date and time with millisecond, UTC)
// - "YYYY-MM-DDThh:mm:ss(+/-)hhmm" (date and time, time zone offset)
// - "YYYY-MM-DDThh:mm:ss.SSS(+/-)hhmm" (date and time with millisecond, time zone offset)
func ParseDatetime(s string) (Datetime, error) {
	var (
		year, month, day, hour, minute, second, milli int
		offset                                        time.Duration
	)

	length := len(s)
	if length < 10 {
		return Datetime{}, fmt.Errorf("%w: string too short", errDatetime)
	}

	// Date: YYYY-MM-DD
	// YYYY is at offset 0
	// MM is at offset 5
	// DD is at offset 8
	// - is at 4 and 7
	// YYYY
	if !unicode.IsDigit(rune(s[0])) || !unicode.IsDigit(rune(s[1])) || !unicode.IsDigit(rune(s[2])) || !unicode.IsDigit(rune(s[3])) {
		return Datetime{}, fmt.Errorf("%w: invalid year", errDatetime)
	}
	year = 1000*int(rune(s[0])-'0') +
		100*int(rune(s[1])-'0') +
		10*int(rune(s[2])-'0') +
		int(rune(s[3])-'0')

	if s[4] != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", errDatetime, strconv.QuoteRune(rune(s[4])))
	}

	// MM
	if !unicode.IsDigit(rune(s[5])) || !unicode.IsDigit(rune(s[6])) {
		return Datetime{}, fmt.Errorf("%w: invalid month", errDatetime)
	}
	month = 10*int(rune(s[5])-'0') + int(rune(s[6])-'0')
	if month > 12 {
		return Datetime{}, fmt.Errorf("%w: month is out of range", errDatetime)
	}

	if s[7] != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", errDatetime, strconv.QuoteRune(rune(s[7])))
	}

	// DD
	if !unicode.IsDigit(rune(s[8])) || !unicode.IsDigit(rune(s[9])) {
		return Datetime{}, fmt.Errorf("%w: invalid day", errDatetime)
	}
	day = 10*int(rune(s[8])-'0') + int(rune(s[9])-'0')
	if day > 31 {
		return Datetime{}, fmt.Errorf("%w: day is out of range", errDatetime)
	}

	// If the length is 10, we only have a date and we're done.
	if length == 10 {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return Datetime{value: t.UnixMilli()}, nil
	}

	// If the length is less than 20, we can't have a valid time.
	if length < 20 {
		return Datetime{}, fmt.Errorf("%w: invalid time", errDatetime)
	}

	// Time: Thh:mm:ss?
	// T is at 10
	// hh is at offset 11
	// mm is at offset 14
	// ss is at offset 17
	// : is at 13 and 16
	// ? is at 19, and... we'll skip to get back to that.

	if s[10] != 'T' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", errDatetime, strconv.QuoteRune(rune(s[10])))
	}

	if !unicode.IsDigit(rune(s[11])) || !unicode.IsDigit(rune(s[12])) {
		return Datetime{}, fmt.Errorf("%w: invalid hour", errDatetime)
	}
	hour = 10*int(rune(s[11])-'0') + int(rune(s[12])-'0')
	if hour > 23 {
		return Datetime{}, fmt.Errorf("%w: hour is out of range", errDatetime)
	}

	if s[13] != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", errDatetime, strconv.QuoteRune(rune(s[13])))
	}

	if !unicode.IsDigit(rune(s[14])) || !unicode.IsDigit(rune(s[15])) {
		return Datetime{}, fmt.Errorf("%w: invalid minute", errDatetime)
	}
	minute = 10*int(rune(s[14])-'0') + int(rune(s[15])-'0')
	if minute > 59 {
		return Datetime{}, fmt.Errorf("%w: minute is out of range", errDatetime)
	}

	if s[16] != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", errDatetime, strconv.QuoteRune(rune(s[16])))
	}

	if !unicode.IsDigit(rune(s[17])) || !unicode.IsDigit(rune(s[18])) {
		return Datetime{}, fmt.Errorf("%w: invalid second", errDatetime)
	}
	second = 10*int(rune(s[17])-'0') + int(rune(s[18])-'0')
	if second > 59 {
		return Datetime{}, fmt.Errorf("%w: second is out of range", errDatetime)
	}

	// At this point, things are variable.
	// 19 can be ., in which case we have milliseconds. (SSS)
	//   ... but we'll still need a Z, or offset. So, we'll introduce
	//       trailerOffset to account for where this starts.
	trailerOffset := 19
	if s[19] == '.' {
		if length < 23 {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", errDatetime)
		}

		if !unicode.IsDigit(rune(s[20])) || !unicode.IsDigit(rune(s[21])) || !unicode.IsDigit(rune(s[22])) {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", errDatetime)
		}

		milli = 100*int(rune(s[20])-'0') + 10*int(rune(s[21])-'0') + int(rune(s[22])-'0')
		trailerOffset = 23
	}

	if length == trailerOffset {
		return Datetime{}, fmt.Errorf("%w: expected time zone designator", errDatetime)
	}

	// At this point, we can only have 2 possible lengths. Anything else is an error.
	switch s[trailerOffset] {
	case 'Z':
		if length > trailerOffset+1 {
			// If something comes after the Z, it's an error
			return Datetime{}, fmt.Errorf("%w: unexpected trailer after time zone designator", errDatetime)
		}
	case '+', '-':
		sign := 1
		if s[trailerOffset] == '-' {
			sign = -1
		}

		if length > trailerOffset+5 {
			return Datetime{}, fmt.Errorf("%w: unexpected trailer after time zone designator", errDatetime)
		} else if length != trailerOffset+5 {
			return Datetime{}, fmt.Errorf("%w: invalid time zone offset", errDatetime)
		}

		// get the time zone offset hhmm.
		if !unicode.IsDigit(rune(s[trailerOffset+1])) || !unicode.IsDigit(rune(s[trailerOffset+2])) || !unicode.IsDigit(rune(s[trailerOffset+3])) || !unicode.IsDigit(rune(s[trailerOffset+4])) {
			return Datetime{}, fmt.Errorf("%w: invalid time zone offset", errDatetime)
		}

		hh := time.Duration(10*int64(rune(s[trailerOffset+1])-'0') + int64(rune(s[trailerOffset+2])-'0'))
		mm := time.Duration(10*int64(rune(s[trailerOffset+3])-'0') + int64(rune(s[trailerOffset+4])-'0'))

		if hh > 23 {
			return Datetime{}, fmt.Errorf("%w: time zone offset hours are out of range", errDatetime)
		}
		if mm > 59 {
			return Datetime{}, fmt.Errorf("%w: time zone offset minutes are out of range", errDatetime)
		}

		offset = time.Duration(sign) * ((hh * time.Hour) + (mm * time.Minute))

	default:
		return Datetime{}, fmt.Errorf("%w: invalid time zone designator", errDatetime)
	}

	t := time.Date(year, time.Month(month), day,
		hour, minute, second,
		int(time.Duration(milli)*time.Millisecond), time.UTC)

	// Don't allow wrapping: https://github.com/cedar-policy/rfcs/pull/94, which can occur
	// because not all months have 31 days, which is our validation range
	_, tmonth, tday := t.Date()
	if time.Month(month) != tmonth || day != tday {
		return Datetime{}, fmt.Errorf("%w: invalid date", errDatetime)
	}

	t = t.Add(offset)
	return Datetime{value: t.UnixMilli()}, nil
}

// Equal returns true if the input represents the same timestamp.
func (d Datetime) Equal(bi Value) bool {
	b, ok := bi.(Datetime)
	return ok && d == b
}

// LessThan returns true if value is less than the argument and they
// are both Datetime values, or an error indicating they aren't
// comparable otherwise
func (d Datetime) LessThan(bi Value) (bool, error) {
	b, ok := bi.(Datetime)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return d.value < b.value, nil
}

// LessThan returns true if value is less than or equal to the
// argument and they are both Datetime values, or an error indicating
// they aren't comparable otherwise
func (d Datetime) LessThanOrEqual(bi Value) (bool, error) {
	b, ok := bi.(Datetime)
	if !ok {
		return false, internal.ErrNotComparable
	}
	return d.value <= b.value, nil
}

// MarshalCedar returns a []byte which, when parsed by the Cedar
// Parser, returns an Equal Datetime value
func (d Datetime) MarshalCedar() []byte {
	return []byte(`datetime("` + d.String() + `")`)
}

// String returns an ISO 8601 millisecond precision timestamp
func (d Datetime) String() string {
	return time.UnixMilli(d.value).UTC().Format("2006-01-02T15:04:05.000Z")
}

// UnmarshalJSON implements encoding/json.Unmarshaler for Datetime
//
// It is capable of unmarshaling 3 different representations supported by Cedar
//   - { "__extn": { "fn": "datetime", "arg": "1970-01-01" }}
//   - { "fn": "datetime", "arg": "1970-01-01" }
//   - "1970-01-01"
func (d *Datetime) UnmarshalJSON(b []byte) error {
	aa, err := unmarshalExtensionValue(b, "datetime", ParseDatetime)
	if err != nil {
		return err
	}

	*d = aa
	return nil
}

// MarshalJSON marshals a Cedar Datetime with the explicit representation
func (d Datetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "datetime",
			Arg: d.String(),
		},
	})
}

// Milliseconds returns the number of milliseconds since the Unix epoch
func (d Datetime) Milliseconds() int64 {
	return d.value
}

// Time returns the UTC time.Time representation of a Datetime.
func (d Datetime) Time() time.Time {
	return time.UnixMilli(d.value).UTC()
}

func (d Datetime) hash() uint64 {
	return uint64(d.value)
}
