package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

type Datetime struct {
	// Value is a timestamp in milliseconds
	Value int64
}

func UnsafeDatetime[T int | int64 | float64](v T) Datetime {
	return Datetime{Value: int64(v)}
}

func FromStdTime(t time.Time) Datetime {
	return Datetime{Value: t.UnixMilli()}
}

// "YYYY-MM-DD" (date only)
// "YYYY-MM-DDThh:mm:ssZ" (UTC)
// "YYYY-MM-DDThh:mm:ss.SSSZ" (UTC with millisecond precision)
// "YYYY-MM-DDThh:mm:ss(+/-)hhmm" (With timezone offset in hours and minutes)
// "YYYY-MM-DDThh:mm:ss.SSS(+/-)hhmm" (With timezone offset in hours and minutes and millisecond precision)
func ParseDatetime(s string) (Datetime, error) {
	var (
		year, month, day, hour, minute, second, milli int
		i                                             int
	)

	length := len(s)

	if length < 10 {
		return Datetime{}, fmt.Errorf("%w: string too short", ErrDatetime)
	}

	// parse YYYY-MM-DD

	// YYYY
	if !(unicode.IsDigit(rune(s[0])) &&
		unicode.IsDigit(rune(s[1])) &&
		unicode.IsDigit(rune(s[2])) &&
		unicode.IsDigit(rune(s[3]))) {
		return Datetime{}, fmt.Errorf("%w: invalid year", ErrDatetime)
	}

	year = 1000*int(rune(s[i])-'0') +
		100*int(rune(s[i+1])-'0') +
		10*int(rune(s[i+2])-'0') +
		int(rune(s[i+3])-'0')

	i = 4

	c := rune(s[i])
	if c != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(c))
	}

	i++

	// MM
	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1]))) {
		return Datetime{}, fmt.Errorf("%w: invalid month", ErrDatetime)
	}

	month = 10*int(rune(s[i])-'0') + int(rune(s[i+1])-'0')

	i = i + 2

	c = rune(s[i])
	if c != '-' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(c))
	}

	i++

	// DD
	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1]))) {
		return Datetime{}, fmt.Errorf("%w: invalid day", ErrDatetime)
	}

	day = 10*int(rune(s[i])-'0') + int(rune(s[i+1])-'0')

	i = i + 2

	if i == length { // done.
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return Datetime{Value: t.UnixMilli()}, nil
	}

	if length < 20 { // can't possibly have a well formed time.
		return Datetime{}, fmt.Errorf("%w: invalid time", ErrDatetime)
	}

	// expect 'T'
	c = rune(s[i])
	if c != 'T' {
		return Datetime{}, fmt.Errorf("%w: expected 'T'", ErrDatetime)
	}

	i++

	// hh
	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1]))) {
		return Datetime{}, fmt.Errorf("%w: invalid hour", ErrDatetime)
	}

	hour = 10*int(rune(s[i])-'0') + int(rune(s[i+1])-'0')

	i = i + 2

	c = rune(s[i])
	if c != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(c))
	}

	i++

	// mm

	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1]))) {
		return Datetime{}, fmt.Errorf("%w: invalid minute", ErrDatetime)
	}

	minute = 10*int(rune(s[i])-'0') + int(rune(s[i+1])-'0')

	i = i + 2

	c = rune(s[i])
	if c != ':' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(c))
	}

	i++

	// ss
	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1]))) {
		return Datetime{}, fmt.Errorf("%w: invalid second", ErrDatetime)
	}

	second = 10*int(rune(s[i])-'0') + int(rune(s[i+1])-'0')

	i += 2

	// .SSS ??

	c = rune(s[i])
	if c == '.' {
		i++

		if i+3 > length {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", ErrDatetime)
		}

		if !(unicode.IsDigit(rune(s[i])) &&
			unicode.IsDigit(rune(s[i+1])) &&
			unicode.IsDigit(rune(s[i+2]))) {
			return Datetime{}, fmt.Errorf("%w: invalid millisecond", ErrDatetime)
		}

		milli = 100*int(rune(s[i])-'0') + 10*int(rune(s[i+1])-'0') + int(rune(s[i+2])-'0')
		i = i + 3
	}

	if i >= length {
		return Datetime{}, fmt.Errorf("%w: expected timezone indicator", ErrDatetime)
	}

	c = rune(s[i])
	if c == 'Z' { // "YYYY-MM-DDThh:mm:ssZ"
		if i+1 == length {
			t := time.Date(year, time.Month(month),
				day, hour, minute, second,
				int(time.Duration(milli)*time.Millisecond), time.UTC)
			return Datetime{Value: t.UnixMilli()}, nil
		}
		return Datetime{}, fmt.Errorf("%w: unexpected trailer after timezone indicator", ErrDatetime)
	}

	// It must have an offset. +/-hhmm

	if i+5 > length {
		return Datetime{}, fmt.Errorf("%w: expected time offset", ErrDatetime)
	} else if i+5 < length {
		return Datetime{}, fmt.Errorf("%w: unexpected trailer", ErrDatetime)
	}

	sign := time.Duration(1)
	c = rune(s[i])
	if c == '-' {
		sign = -sign
	} else if c != '+' {
		return Datetime{}, fmt.Errorf("%w: unexpected character %s", ErrDatetime, strconv.QuoteRune(c))
	}

	i++

	if !(unicode.IsDigit(rune(s[i])) &&
		unicode.IsDigit(rune(s[i+1])) &&
		unicode.IsDigit(rune(s[i+2])) &&
		unicode.IsDigit(rune(s[i+3]))) {
		return Datetime{}, fmt.Errorf("%w: invalid time offset", ErrDatetime)
	}

	offsetH := time.Duration(10*int64(rune(s[i])-'0')+int64(rune(s[i+1])-'0')) * time.Hour
	offsetM := time.Duration(10*int64(rune(s[i+2])-'0')+int64(rune(s[i+3])-'0')) * time.Minute

	t := time.Date(year, time.Month(month), day,
		hour, minute, second,
		int(time.Duration(milli)*time.Millisecond), time.UTC)
	t = t.Add(sign * (offsetH + offsetM))
	return Datetime{Value: t.UnixMilli()}, nil
}

func (a Datetime) Equal(bi Value) bool {
	b, ok := bi.(Datetime)
	return ok && a == b
}

func (a Datetime) Less(bi Value) bool {
	b, ok := bi.(Datetime)
	return ok && a.Value < b.Value
}

func (a Datetime) LessEqual(bi Value) bool {
	b, ok := bi.(Datetime)
	return ok && a.Value <= b.Value
}

func (a Datetime) MarshalCedar() []byte {
	return []byte(`datetime("` + a.String() + `")`)
}

func (a Datetime) String() string {
	return time.UnixMilli(a.Value).UTC().Format("2006-01-02T15:04:05.000Z")
}

func (a *Datetime) UnmarshalJSON(b []byte) error {
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		// NOTE: cedar supports two other forms, for now we're only supporting the smallest implicit and explicit form.
		// The following are not supported:
		// "datetime(\"1970-01-01T00:00:00Z\")"
		// {"fn":"datetime","arg":"1970-01-01T00:00:00Z"}
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			return errJSONExtNotFound
		}
		if res.Extn.Fn != "datetime" {
			return errJSONExtFnMatch
		}
		arg = res.Extn.Arg
	}
	aa, err := ParseDatetime(arg)
	if err != nil {
		return err
	}
	*a = aa
	return nil
}

func (a Datetime) MarshalJSON() ([]byte, error) {
	return []byte(`datetime("` + a.String() + `")`), nil
}

func (a Datetime) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "datetime",
			Arg: a.String(),
		},
	})
}

func (v Datetime) deepClone() Value { return v }
