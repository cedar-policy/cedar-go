package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

// Datetime represents a Cedar datetime value
type Datetime struct {
	// value is a timestamp in milliseconds
	value int64
}

// FromStdTime returns a Cedar Datetime from a Go time.Time value
func FromStdTime(t time.Time) Datetime {
	return Datetime{value: t.UnixMilli()}
}

// DatetimeFromMillis returns a Datetime from milliseconds
func DatetimeFromMillis(ms int64) Datetime {
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
		return Datetime{}, fmt.Errorf("%w: string too short", ErrDatetime)
	}

	// Date: YYYY-MM-DD
	// YYYY is at offset 0
	// MM is at offset 5
	// DD is at offset 8
	// - is at 4 and 7
	// YYYY
	if !(unicode.IsDigit(rune(s[0])) &&
		unicode.IsDigit(rune(s[1])) &&
		unicode.IsDigit(rune(s[2])) &&
		unicode.IsDigit(rune(s[3]))) {
		return Datetime{}, fmt.Errorf("%w: invalid year", ErrDatetime)
	}
	year = 1000*int(rune(s[0])-'0') +
		100*int(rune(s[1])-'0') +
		10*int(rune(s[2])-'0') +
		int(rune(s[3])-'0')

	if s[4] != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(rune(s[4])))
	}

	// MM
	if !(unicode.IsDigit(rune(s[5])) &&
		unicode.IsDigit(rune(s[6]))) {
		return Datetime{}, fmt.Errorf("%w: invalid month", ErrDatetime)
	}
	month = 10*int(rune(s[5])-'0') + int(rune(s[6])-'0')

	if s[7] != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(rune(s[7])))
	}

	// DD
	if !(unicode.IsDigit(rune(s[8])) &&
		unicode.IsDigit(rune(s[9]))) {
		return Datetime{}, fmt.Errorf("%w: invalid day", ErrDatetime)
	}
	day = 10*int(rune(s[8])-'0') + int(rune(s[9])-'0')

	// If the length is 10, we only have a date and we're done.
	if length == 10 {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return Datetime{value: t.UnixMilli()}, nil
	}

	// If the length is less than 20, we can't have a valid time.
	if length < 20 {
		return Datetime{}, fmt.Errorf("%w: invalid time", ErrDatetime)
	}

	// Time: Thh:mm:ss?
	// T is at 10
	// hh is at offset 11
	// mm is at offset 14
	// ss is at offset 17
	// : is at 13 and 16
	// ? is at 19, and... we'll skip to get back to that.

	if s[10] != 'T' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(rune(s[10])))
	}

	if !(unicode.IsDigit(rune(s[11])) &&
		unicode.IsDigit(rune(s[12]))) {
		return Datetime{}, fmt.Errorf("%w: invalid hour", ErrDatetime)
	}
	hour = 10*int(rune(s[11])-'0') + int(rune(s[12])-'0')

	if s[13] != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(rune(s[13])))
	}

	if !(unicode.IsDigit(rune(s[14])) &&
		unicode.IsDigit(rune(s[15]))) {
		return Datetime{}, fmt.Errorf("%w: invalid minute", ErrDatetime)
	}
	minute = 10*int(rune(s[14])-'0') + int(rune(s[15])-'0')

	if s[16] != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(rune(s[16])))
	}

	if !(unicode.IsDigit(rune(s[17])) &&
		unicode.IsDigit(rune(s[18]))) {
		return Datetime{}, fmt.Errorf("%w: invalid second", ErrDatetime)
	}
	second = 10*int(rune(s[17])-'0') + int(rune(s[18])-'0')

	// At this point, things are variable.
	// 19 can be ., in which case we have milliseconds. (SSS)
	//   ... but we'll still need a Z, or offset. So, we'll introduce
	//       trailerOffset to account for where this starts.
	trailerOffset := 19
	if s[19] == '.' {
		if length < 23 {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", ErrDatetime)
		}

		if !(unicode.IsDigit(rune(s[20])) &&
			unicode.IsDigit(rune(s[21])) &&
			unicode.IsDigit(rune(s[22]))) {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", ErrDatetime)
		}

		milli = 100*int(rune(s[20])-'0') + 10*int(rune(s[21])-'0') + int(rune(s[22])-'0')
		trailerOffset = 23
	}

	if length == trailerOffset {
		return Datetime{}, fmt.Errorf("%w: expected time zone designator", ErrDatetime)
	}

	// At this point, we can only have 2 possible lengths. Anything else is an error.
	switch s[trailerOffset] {
	case 'Z':
		if length > trailerOffset+1 {
			// If something comes after the Z, it's an error
			return Datetime{}, fmt.Errorf("%w: unexpected trailer after time zone designator", ErrDatetime)
		}
	case '+', '-':
		sign := 1
		if s[trailerOffset] == '-' {
			sign = -1
		}

		if length > trailerOffset+5 {
			return Datetime{}, fmt.Errorf("%w: unexpected trailer after time zone designator", ErrDatetime)
		} else if length != trailerOffset+5 {
			return Datetime{}, fmt.Errorf("%w: invalid time zone offset", ErrDatetime)
		}

		// get the time zone offset hhmm.
		if !(unicode.IsDigit(rune(s[trailerOffset+1])) &&
			unicode.IsDigit(rune(s[trailerOffset+2])) &&
			unicode.IsDigit(rune(s[trailerOffset+3])) &&
			unicode.IsDigit(rune(s[trailerOffset+4]))) {
			return Datetime{}, fmt.Errorf("%w: invalid time zone offset", ErrDatetime)
		}

		hh := time.Duration(10*int64(rune(s[trailerOffset+1])-'0')+int64(rune(s[trailerOffset+2])-'0')) * time.Hour
		mm := time.Duration(10*int64(rune(s[trailerOffset+3])-'0')+int64(rune(s[trailerOffset+4])-'0')) * time.Minute
		offset = time.Duration(sign) * (hh + mm)

	default:
		return Datetime{}, fmt.Errorf("%w: invalid time zone designator", ErrDatetime)
	}

	t := time.Date(year, time.Month(month), day,
		hour, minute, second,
		int(time.Duration(milli)*time.Millisecond), time.UTC)
	t = t.Add(offset)
	return Datetime{value: t.UnixMilli()}, nil
}

// Equal returns true if the input represents the same timestamp.
func (a Datetime) Equal(bi Value) bool {
	b, ok := bi.(Datetime)
	return ok && a == b
}

// LessThan returns true if value is less than the argument and they
// are both Datetime values, or an error indicating they aren't
// comparable otherwise
func (a Datetime) LessThan(bi Value) (bool, error) {
	b, ok := bi.(Datetime)
	if !ok {
		return false, ErrNotComparable
	}
	return a.value < b.value, nil
}

// LessThan returns true if value is less than or equal to the
// argument and they are both Datetime values, or an error indicating
// they aren't comparable otherwise
func (a Datetime) LessThanOrEqual(bi Value) (bool, error) {
	b, ok := bi.(Datetime)
	if !ok {
		return false, ErrNotComparable
	}
	return a.value <= b.value, nil
}

// MarshalCedar returns a []byte which, when parsed by the Cedar
// Parser, returns an Equal Datetime value
func (a Datetime) MarshalCedar() []byte {
	return []byte(`datetime("` + a.String() + `")`)
}

// String returns an ISO 8601 millisecond precision timestamp
func (a Datetime) String() string {
	return time.UnixMilli(a.value).UTC().Format("2006-01-02T15:04:05.000Z")
}

// UnmarshalJSON implements encoding/json.Unmarshaler for Datetime
//
// It is capable of unmarshaling 4 different representations supported by Cedar
// - { "__extn": { "fn": "datetime", "arg": "1970-01-01" }}
// - { "fn": "datetime", "arg": "1970-01-01" }
// - "datetime(\"1970-01-01\")"
// - "1970-01-01"
func (a *Datetime) UnmarshalJSON(b []byte) error {
	var arg string
	if bytes.HasPrefix(b, []byte(`"datetime(\"`)) && bytes.HasSuffix(b, []byte(`\")"`)) {
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
			if res2.Fn != "datetime" {
				return errJSONExtFnMatch
			}
			arg = res2.Arg
		} else if res.Extn.Fn != "datetime" {
			return errJSONExtFnMatch
		} else {
			arg = res.Extn.Arg
		}
	}
	aa, err := ParseDatetime(arg)
	if err != nil {
		return err
	}
	*a = aa
	return nil
}

// MarshalJSON implements the encoding/json.Marshaler interface
//
// It produces the direct representation of a Cedar Datetime.
func (a Datetime) MarshalJSON() ([]byte, error) {
	return []byte(`datetime("` + a.String() + `")`), nil
}

// ExplicitMarshalJSON Marshal's a Cedar Datetime with the explicit
// representation
func (a Datetime) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "datetime",
			Arg: a.String(),
		},
	})
}

// Milliseconds returns the number of milliseconds since the Unix epoch
func (a Datetime) Milliseconds() int64 {
	return a.value
}

func (v Datetime) hash() uint64 {
	return uint64(v.value)
}
