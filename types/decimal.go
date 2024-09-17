package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

// A Decimal is a value with both a whole number part and a decimal part of no
// more than four digits. In Go this is stored as an int64, the precision is
// defined by the constant DecimalPrecision.
type Decimal struct {
	Value int64
}

// UnsafeDecimal creates a decimal via unsafe conversion from int, int64, float64.
// Precision may be lost and overflows may occur.
func UnsafeDecimal[T int | int64 | float64](v T) Decimal {
	return Decimal{Value: int64(v * DecimalPrecision)}
}

// DecimalPrecision is the precision of a Decimal.
const DecimalPrecision = 10000

// ParseDecimal takes a string representation of a decimal number and converts it into a Decimal type.
func ParseDecimal(s string) (Decimal, error) {
	// Check for empty string.
	if len(s) == 0 {
		return Decimal{}, fmt.Errorf("%w: string too short", ErrDecimal)
	}
	i := 0

	// Parse an optional '-'.
	negative := false
	if s[i] == '-' {
		negative = true
		i++
		if i == len(s) {
			return Decimal{}, fmt.Errorf("%w: string too short", ErrDecimal)
		}
	}

	// Parse the required first digit.
	c := rune(s[i])
	if !unicode.IsDigit(c) {
		return Decimal{}, fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
	}
	integer := int64(c - '0')
	i++

	// Parse any other digits, ending with i pointing to '.'.
	for ; ; i++ {
		if i == len(s) {
			return Decimal{}, fmt.Errorf("%w: string missing decimal point", ErrDecimal)
		}
		c = rune(s[i])
		if c == '.' {
			break
		}
		if !unicode.IsDigit(c) {
			return Decimal{}, fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
		}
		integer = 10*integer + int64(c-'0')
		if integer > 922337203685477 {
			return Decimal{}, fmt.Errorf("%w: overflow", ErrDecimal)
		}
	}

	// Advance past the '.'.
	i++

	// Parse the fraction part
	fraction := int64(0)
	fractionDigits := 0
	for ; i < len(s); i++ {
		c = rune(s[i])
		if !unicode.IsDigit(c) {
			return Decimal{}, fmt.Errorf("%w: unexpected character %s", ErrDecimal, strconv.QuoteRune(c))
		}
		fraction = 10*fraction + int64(c-'0')
		fractionDigits++
	}

	// Adjust the fraction part based on how many digits we parsed.
	switch fractionDigits {
	case 0:
		return Decimal{}, fmt.Errorf("%w: missing digits after decimal point", ErrDecimal)
	case 1:
		fraction *= 1000
	case 2:
		fraction *= 100
	case 3:
		fraction *= 10
	case 4:
	default:
		return Decimal{}, fmt.Errorf("%w: too many digits after decimal point", ErrDecimal)
	}

	// Check for overflow before we put the number together.
	if integer >= 922337203685477 && (fraction > 5808 || (!negative && fraction == 5808)) {
		return Decimal{}, fmt.Errorf("%w: overflow", ErrDecimal)
	}

	// Put the number together.
	if negative {
		// Doing things in this order keeps us from overflowing when parsing
		// -922337203685477.5808. This isn't technically necessary because the
		// go spec defines arithmetic to be well-defined when overflowing.
		// However, doing things this way doesn't hurt, so let's be pedantic.
		return Decimal{Value: DecimalPrecision*-integer - fraction}, nil
	} else {
		return Decimal{Value: DecimalPrecision*integer + fraction}, nil
	}
}

func (a Decimal) Equal(bi Value) bool {
	b, ok := bi.(Decimal)
	return ok && a == b
}

// MarshalCedar produces a valid MarshalCedar language representation of the Decimal, e.g. `decimal("12.34")`.
func (v Decimal) MarshalCedar() []byte { return []byte(`decimal("` + v.String() + `")`) }

// String produces a string representation of the Decimal, e.g. `12.34`.
func (v Decimal) String() string {
	var res string
	if v.Value < 0 {
		// Make sure we don't overflow here. Also, go truncates towards zero.
		integer := v.Value / DecimalPrecision
		decimal := integer*DecimalPrecision - v.Value
		res = fmt.Sprintf("-%d.%04d", -integer, decimal)
	} else {
		res = fmt.Sprintf("%d.%04d", v.Value/DecimalPrecision, v.Value%DecimalPrecision)
	}

	// Trim off up to three trailing zeros.
	right := len(res)
	for trimmed := 0; right-1 >= 0 && trimmed < 3; right, trimmed = right-1, trimmed+1 {
		if res[right-1] != '0' {
			break
		}
	}
	return res[:right]
}

func (v *Decimal) UnmarshalJSON(b []byte) error {
	var arg string
	if len(b) > 0 && b[0] == '"' {
		if err := json.Unmarshal(b, &arg); err != nil {
			return errors.Join(errJSONDecode, err)
		}
	} else {
		// NOTE: cedar supports two other forms, for now we're only supporting the smallest implicit and explicit form.
		// The following are not supported:
		// "decimal(\"1234.5678\")"
		// {"fn":"decimal","arg":"1234.5678"}
		var res extValueJSON
		if err := json.Unmarshal(b, &res); err != nil {
			return errors.Join(errJSONDecode, err)
		}
		if res.Extn == nil {
			return errJSONExtNotFound
		}
		if res.Extn.Fn != "decimal" {
			return errJSONExtFnMatch
		}
		arg = res.Extn.Arg
	}
	vv, err := ParseDecimal(arg)
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

// ExplicitMarshalJSON marshals the Decimal into JSON using the implicit form.
func (v Decimal) MarshalJSON() ([]byte, error) { return []byte(`"` + v.String() + `"`), nil }

// ExplicitMarshalJSON marshals the Decimal into JSON using the explicit form.
func (v Decimal) ExplicitMarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "decimal",
			Arg: v.String(),
		},
	})
}

func (v Decimal) hash() uint64 {
	return uint64(v.Value)
}
