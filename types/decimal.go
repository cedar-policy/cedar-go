package types

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cedar-policy/cedar-go/internal"
	"golang.org/x/exp/constraints"
)

var errDecimal = internal.ErrDecimal

// decimalPrecision is the precision of a Decimal.
const decimalPrecision = 10000

// A Decimal is a value with both a whole number part and a decimal part of no
// more than four digits. A decimal value can range from -922337203685477.5808
// to 922337203685477.5807.
type Decimal struct {
	value int64
}

// newDecimal returns a Decimal value of the form intPart.tenThousandths. The
// sign of intPart and tenThousandths should match.
func newDecimal(intPart int64, tenThousandths int16) (Decimal, error) {
	if intPart > 922337203685477 || (intPart == 922337203685477 && tenThousandths > 5807) {
		return Decimal{}, fmt.Errorf("%w: value would overflow", errDecimal)
	} else if intPart < -922337203685477 || (intPart == -922337203685477 && tenThousandths < -5808) {
		return Decimal{}, fmt.Errorf("%w: value would underflow", errDecimal)
	}

	return Decimal{value: intPart*decimalPrecision + int64(tenThousandths)}, nil
}

// NewDecimal returns a Decimal value of i * 10^exponent.
func NewDecimal(i int64, exponent int) (Decimal, error) {
	if exponent < -4 || exponent > 14 {
		return Decimal{}, fmt.Errorf("%w: exponent value of %v exceeds maximum range of Decimal", errDecimal, exponent)
	}

	var intPart int64
	var fracPart int64
	if exponent <= 0 {
		intPart = i / int64(math.Pow10(-exponent))
		fracPart = i % int64(math.Pow10(-exponent)) * int64(math.Pow10(4+exponent))
	} else {
		intPart = i * int64(math.Pow10(exponent))
		if i > 0 && intPart < i {
			return Decimal{}, fmt.Errorf("%w: value %ve%v would overflow", errDecimal, i, exponent)
		} else if i < 0 && intPart > i {
			return Decimal{}, fmt.Errorf("%w: value %ve%v would underflow", errDecimal, i, exponent)
		}
	}

	return newDecimal(intPart, int16(fracPart))
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
		return Decimal{}, fmt.Errorf("%w: value %v would overflow", errDecimal, f)
	} else if f < math.MinInt64 {
		return Decimal{}, fmt.Errorf("%w: value %v would underflow", errDecimal, f)
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
	decimalIndex := strings.Index(s, ".")
	if decimalIndex < 0 {
		return Decimal{}, fmt.Errorf("%w: missing decimal point", errDecimal)
	}

	intPart, err := strconv.ParseInt(s[0:decimalIndex], 10, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return Decimal{}, fmt.Errorf("%w: value would overflow", errDecimal)
		}
		return Decimal{}, fmt.Errorf("%w: %w", errDecimal, err)
	}

	fracPartStr := s[decimalIndex+1:]
	fracPart, err := strconv.ParseUint(fracPartStr, 10, 16)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return Decimal{}, fmt.Errorf("%w: fractional part exceeds Decimal precision", errDecimal)
		}
		return Decimal{}, fmt.Errorf("%w: %w", errDecimal, err)
	}

	decimalPlaces := len(fracPartStr)
	if decimalPlaces > 4 {
		return Decimal{}, fmt.Errorf("%w: fractional part exceeds Decimal precision", errDecimal)
	}

	tenThousandths := int16(fracPart) * int16(math.Pow10(4-decimalPlaces))
	if intPart < 0 {
		tenThousandths = -tenThousandths
	}

	return newDecimal(intPart, tenThousandths)
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

// UnmarshalJSON implements encoding/json.Unmarshaler for Decimal
//
// It is capable of unmarshaling 3 different representations supported by Cedar
//   - { "__extn": { "fn": "decimal", "arg": "1234.5678" }}
//   - { "fn": "decimal", "arg": "1234.5678" }
//   - "1234.5678"
func (d *Decimal) UnmarshalJSON(b []byte) error {
	dd, err := unmarshalExtensionValue(b, "decimal", ParseDecimal)
	if err != nil {
		return err
	}

	*d = dd
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

// Float returns a float64 representation of a Decimal.  Warning: some precision
// may be lost during this conversion.
func (d Decimal) Float() float64 {
	return float64(d.value) / decimalPrecision
}

func (d Decimal) hash() uint64 {
	return uint64(d.value)
}
