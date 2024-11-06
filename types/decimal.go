package types

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"unicode"

	"golang.org/x/exp/constraints"
)

// decimalPrecision is the precision of a Decimal.
const decimalPrecision = 10000

var DecimalMax = Decimal{value: math.MaxInt64}
var DecimalMin = Decimal{value: math.MinInt64}

// A Decimal is a value with both a whole number part and a decimal part of no
// more than four digits. A decimal value can range from -922337203685477.5808
// to 922337203685477.5807.
type Decimal struct {
	value int64
}

// NewDecimal returns a Decimal value of i * 10^exponent.
func NewDecimal(i int64, exponent int) (Decimal, error) {
	if exponent < -4 || exponent > 14 {
		return Decimal{}, fmt.Errorf("%w: exponent value of %v exceeds maximum range of Decimal", ErrDecimal, exponent)
	}

	var intPart int64
	var fracPart int64
	if exponent <= 0 {
		intPart = i / int64(math.Pow10(-exponent))
		fracPart = i % int64(math.Pow10(-exponent)) * int64(math.Pow10(4+exponent))
	} else {
		intPart = i * int64(math.Pow10(exponent))
		if i > 0 && intPart < i {
			return Decimal{}, fmt.Errorf("%w: value %ve%v would overflow", ErrDecimal, i, exponent)
		} else if i < 0 && intPart > i {
			return Decimal{}, fmt.Errorf("%w: value %ve%v would underflow", ErrDecimal, i, exponent)
		}
	}

	if intPart > 922337203685477 || (intPart == 922337203685477 && fracPart > 5807) {
		return Decimal{}, fmt.Errorf("%w: value %ve%v would overflow", ErrDecimal, i, exponent)
	} else if intPart < -922337203685477 || (intPart == -922337203685477 && fracPart < -5808) {
		return Decimal{}, fmt.Errorf("%w: value %ve%v would underflow", ErrDecimal, i, exponent)
	}

	return Decimal{value: intPart*decimalPrecision + fracPart}, nil
}

// NewDecimalFromInt returns a Decimal with the whole integer value provided
func NewDecimalFromInt[T constraints.Signed](i T) (Decimal, error) {
	return NewDecimal(int64(i), 0)
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
	f = f * decimalPrecision
	if f > math.MaxInt64 {
		return Decimal{}, fmt.Errorf("%w: value %v would overflow", ErrDecimal, f)
	} else if f < math.MinInt64 {
		return Decimal{}, fmt.Errorf("%w: value %v would underflow", ErrDecimal, f)
	}

	return Decimal{int64(f)}, nil
}

// Compare returns
//
//	-1 if d is less than other,
//	 0 if d equals other,
//	+1 if d is greater than other.
func (d Decimal) Compare(other Decimal) int {
	return cmp.Compare(d.value, other.value)
}

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
		return Decimal{value: decimalPrecision*-integer - fraction}, nil
	} else {
		return Decimal{value: decimalPrecision*integer + fraction}, nil
	}
}

func (d Decimal) Equal(bi Value) bool {
	b, ok := bi.(Decimal)
	return ok && d == b
}

// MarshalCedar produces a valid MarshalCedar language representation of the Decimal, e.g. `decimal("12.34")`.
func (d Decimal) MarshalCedar() []byte { return []byte(`decimal("` + d.String() + `")`) }

// String produces a string representation of the Decimal, e.g. `12.34`.
func (d Decimal) String() string {
	var res string
	if d.value < 0 {
		// Make sure we don't overflow here. Also, go truncates towards zero.
		integer := d.value / decimalPrecision
		decimal := integer*decimalPrecision - d.value
		res = fmt.Sprintf("-%d.%04d", -integer, decimal)
	} else {
		res = fmt.Sprintf("%d.%04d", d.value/decimalPrecision, d.value%decimalPrecision)
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

func (d *Decimal) UnmarshalJSON(b []byte) error {
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
	*d = vv
	return nil
}

// MarshalJSON marshals the Decimal into JSON using the explicit form.
func (d Decimal) MarshalJSON() ([]byte, error) {
	return json.Marshal(extValueJSON{
		Extn: &extn{
			Fn:  "decimal",
			Arg: d.String(),
		},
	})
}

func (d Decimal) hash() uint64 {
	return uint64(d.value)
}
