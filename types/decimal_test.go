package types_test

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestDecimal(t *testing.T) {
	t.Parallel()
	{
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
	}

	{
		tests := []struct{ in, errStr string }{
			{"", "error parsing decimal value: string too short"},
			{"-", "error parsing decimal value: string too short"},
			{"a", "error parsing decimal value: unexpected character 'a'"},
			{"-a", "error parsing decimal value: unexpected character 'a'"},
			{"'", `error parsing decimal value: unexpected character '\''`},
			{`-\\`, `error parsing decimal value: unexpected character '\\'`},
			{"0", "error parsing decimal value: string missing decimal point"},
			{"1a", "error parsing decimal value: unexpected character 'a'"},
			{"1a", "error parsing decimal value: unexpected character 'a'"},
			{"1.", "error parsing decimal value: missing digits after decimal point"},
			{"1.00000", "error parsing decimal value: too many digits after decimal point"},
			{"1.a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0000a", "error parsing decimal value: unexpected character 'a'"},
			{"1.0000a", "error parsing decimal value: unexpected character 'a'"},

			{"1000000000000000.0", "error parsing decimal value: overflow"},
			{"-1000000000000000.0", "error parsing decimal value: overflow"},
			{"922337203685477.5808", "error parsing decimal value: overflow"},
			{"922337203685478.0", "error parsing decimal value: overflow"},
			{"-922337203685477.5809", "error parsing decimal value: overflow"},
			{"-922337203685478.0", "error parsing decimal value: overflow"},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%s->%s", tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDecimal(tt.in)
				testutil.AssertError(t, err, types.ErrDecimal)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.UnsafeDecimal(1)
		one2 := types.UnsafeDecimal(1)
		zero := types.UnsafeDecimal(0)
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

}
