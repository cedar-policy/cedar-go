package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestDecimal(t *testing.T) {
	t.Parallel()
	t.Run("ParseDecimal", func(t *testing.T) {
		tests := []struct{ in, out string }{
			{"1.2345", "1.2345"},
			{"1.2340", "1.234"},
			{"1.2300", "1.23"},
			{"1.2000", "1.2"},
			{"1.0000", "1.0"},
			{"1.234", "1.234"},
			{"1.230", "1.23"},
			{"1.200", "1.2"},
			{"1.000", "1.0"},
			{"1.23", "1.23"},
			{"1.20", "1.2"},
			{"1.00", "1.0"},
			{"1.2", "1.2"},
			{"1.0", "1.0"},
			{"01.0100", "1.01"},
			{"01.2345", "1.2345"},
			{"01.2340", "1.234"},
			{"01.2300", "1.23"},
			{"01.2000", "1.2"},
			{"01.0000", "1.0"},
			{"01.234", "1.234"},
			{"01.230", "1.23"},
			{"01.200", "1.2"},
			{"01.000", "1.0"},
			{"01.23", "1.23"},
			{"01.20", "1.2"},
			{"01.00", "1.0"},
			{"01.2", "1.2"},
			{"01.0", "1.0"},
			{"1234.5678", "1234.5678"},
			{"1234.5670", "1234.567"},
			{"1234.5600", "1234.56"},
			{"1234.5000", "1234.5"},
			{"1234.0000", "1234.0"},
			{"1234.567", "1234.567"},
			{"1234.560", "1234.56"},
			{"1234.500", "1234.5"},
			{"1234.000", "1234.0"},
			{"1234.56", "1234.56"},
			{"1234.50", "1234.5"},
			{"1234.00", "1234.0"},
			{"1234.5", "1234.5"},
			{"1234.0", "1234.0"},
			{"0.0", "0.0"},
			{"00.0", "0.0"},
			{"000000000000000000000000000000000000000000000000000000000000000000.0", "0.0"},
			{"922337203685477.5807", "922337203685477.5807"},
			{"-922337203685477.5808", "-922337203685477.5808"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s->%s", tt.in, tt.out), func(t *testing.T) {
				t.Parallel()
				d, err := types.ParseDecimal(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.out)
			})
		}
	})

	t.Run("ParseDecimalErrors", func(t *testing.T) {
		tests := []struct{ in, errStr string }{
			{"", "error parsing decimal value: missing decimal point"},
			{"-", "error parsing decimal value: missing decimal point"},
			{"a", "error parsing decimal value: missing decimal point"},
			{"0", "error parsing decimal value: missing decimal point"},
			{"-a.", `error parsing decimal value: strconv.ParseInt: parsing "-a": invalid syntax`},
			{"'.", `error parsing decimal value: strconv.ParseInt: parsing "'": invalid syntax`},
			{"-\\.", `error parsing decimal value: strconv.ParseInt: parsing "-\\": invalid syntax`},
			{"1a.0", `error parsing decimal value: strconv.ParseInt: parsing "1a": invalid syntax`},
			{"1.", `error parsing decimal value: strconv.ParseUint: parsing "": invalid syntax`},
			{"1.00000", "error parsing decimal value: fractional part exceeds Decimal precision"},
			{"1.12345", "error parsing decimal value: fractional part exceeds Decimal precision"},
			{"1.1234567890", "error parsing decimal value: fractional part exceeds Decimal precision"},
			{"1.a", `error parsing decimal value: strconv.ParseUint: parsing "a": invalid syntax`},
			{"1.0a", `error parsing decimal value: strconv.ParseUint: parsing "0a": invalid syntax`},
			{"1.0000a", `error parsing decimal value: strconv.ParseUint: parsing "0000a": invalid syntax`},

			{"10000000000000000000.0", "error parsing decimal value: value would overflow"},
			{"1000000000000000.0", "error parsing decimal value: value would overflow"},
			{"-1000000000000000.0", "error parsing decimal value: value would underflow"},
			{"922337203685477.5808", "error parsing decimal value: value would overflow"},
			{"922337203685478.0", "error parsing decimal value: value would overflow"},
			{"-922337203685477.5809", "error parsing decimal value: value would underflow"},
			{"-922337203685478.0", "error parsing decimal value: value would underflow"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s->%s", tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDecimal(tt.in)
				testutil.ErrorIs(t, err, internal.ErrDecimal)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := testutil.Must(types.NewDecimal(1, 0))
		one2 := testutil.Must(types.NewDecimal(1, 0))
		zero := testutil.Must(types.NewDecimal(0, 0))
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)

		maxVal := testutil.Must(types.NewDecimal(9223372036854775807, -4))
		testutil.Equals(t, maxVal, testutil.Must(types.ParseDecimal("922337203685477.5807")))
		minVal := testutil.Must(types.NewDecimal(-9223372036854775808, -4))
		testutil.Equals(t, minVal, testutil.Must(types.ParseDecimal("-922337203685477.5808")))
	})

	t.Run("Compare", func(t *testing.T) {
		t.Parallel()
		one := testutil.Must(types.NewDecimal(1, 0))
		zero := testutil.Must(types.NewDecimal(0, 0))
		testutil.Equals(t, one.Compare(zero), 1)
		testutil.Equals(t, one.Compare(one), 0)
		testutil.Equals(t, zero.Compare(one), -1)
	})

	t.Run("NewDecimal", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in   int64
			exp  int
			want string
		}{
			{9223372036854775807, -4, "922337203685477.5807"},
			{922337203685477580, -3, "922337203685477.58"},
			{92233720368547758, -2, "922337203685477.58"},
			{9223372036854775, -1, "922337203685477.5"},
			{922337203685477, 0, "922337203685477.0"},
			{-9223372036854775808, -4, "-922337203685477.5808"},
			{-922337203685477580, -3, "-922337203685477.58"},
			{-92233720368547758, -2, "-922337203685477.58"},
			{-9223372036854775, -1, "-922337203685477.5"},
			{-922337203685477, 0, "-922337203685477.0"},
			{92233720368547, 1, "922337203685470.0"},
			{9223372036854, 2, "922337203685400.0"},
			{922337203685, 3, "922337203685000.0"},
			{9, 14, "900000000000000.0"},
		}
		for _, tt := range tests {
			t.Run(tt.want, func(t *testing.T) {
				t.Parallel()
				d, err := types.NewDecimal(tt.in, tt.exp)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.want)
			})
		}
	})

	t.Run("NewDecimalFromInt", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in   int64
			want string
		}{
			{0, "0.0"},
			{1, "1.0"},
			{-1, "-1.0"},
			{922337203685477, "922337203685477.0"},
			{-922337203685477, "-922337203685477.0"},
		}
		for _, tt := range tests {
			t.Run(tt.want, func(t *testing.T) {
				t.Parallel()
				d, err := types.NewDecimalFromInt(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.want)
			})
		}
	})
	t.Run("NewDecimalOverflow", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in  int64
			exp int
		}{
			{922337203685477581, -3},
			{92233720368547759, -2},
			{9223372036854776, -1},
			{922337203685478, 0},
			{92233720368548, 1},
			{922337203685477581, 2},
			{10, 14},
			{1, 15},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%ve%v", tt.in, tt.exp), func(t *testing.T) {
				t.Parallel()
				_, err := types.NewDecimal(tt.in, tt.exp)
				testutil.ErrorIs(t, err, internal.ErrDecimal)
			})
		}
	})

	t.Run("NewDecimalUnderflow", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in  int64
			exp int
		}{
			{-922337203685477581, -3},
			{-92233720368547759, -2},
			{-9223372036854776, -1},
			{-922337203685478, 0},
			{-92233720368548, 1},
			{-922337203685477581, 2},
			{-10, 14},
			{-1, 15},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%ve%v", tt.in, tt.exp), func(t *testing.T) {
				t.Parallel()
				_, err := types.NewDecimal(tt.in, tt.exp)
				testutil.ErrorIs(t, err, internal.ErrDecimal)
			})
		}
	})

	t.Run("NewDecimalFromFloat", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in   float64
			want string
		}{
			{0.0, "0.0"},
			{1.0, "1.0"},
			{-1.0, "-1.0"},
			{1.23451, "1.2345"},
			{1.23456, "1.2345"},
			{12345678901.2345, "12345678901.2345"},
			{123456789012.3456, "123456789012.3456"},
		}
		for _, tt := range tests {
			t.Run(tt.want, func(t *testing.T) {
				t.Parallel()
				d, err := types.NewDecimalFromFloat(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.want)
			})
		}
	})

	t.Run("NewDecimalFromFloatOverflow", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			in float64
		}{
			{922337203685477.6875},
			{-922337203685477.6876},
			{1000000000000000.0},
			{-1000000000000000.0},
		}
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
				_, err := types.NewDecimalFromFloat(tt.in)
				testutil.ErrorIs(t, err, internal.ErrDecimal)
			})
		}
	})

	t.Run("Float", func(t *testing.T) {
		t.Parallel()
		in, err := types.NewDecimalFromFloat(42.42)
		testutil.OK(t, err)
		got := in.Float()
		want := 42.42
		testutil.Equals(t, got, want)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(
			t,
			string(testutil.Must(types.NewDecimal(42, 0)).MarshalCedar()),
			`decimal("42.0")`)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		expected := `{
			"__extn": {
				"fn": "decimal",
				"arg": "1234.5678"
			}
		}`
		d1 := testutil.Must(types.NewDecimal(12345678, -4))
		testutil.JSONMarshalsTo(t, d1, expected)

		var d2 types.Decimal
		err := json.Unmarshal([]byte(expected), &d2)
		testutil.OK(t, err)
		testutil.Equals(t, d1, d2)
	})

	t.Run("UnmarshalJSON/error", func(t *testing.T) {
		t.Parallel()
		var dt2 types.Decimal
		err := json.Unmarshal([]byte("{}"), &dt2)
		testutil.Error(t, err)
	})
}
