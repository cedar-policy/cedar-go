package types_test

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestDatetime(t *testing.T) {
	t.Parallel()
	{
		tests := []struct{ in, out string }{
			// date only YYYY-MM-DD
			{"1970-01-01", "1970-01-01T00:00:00.000Z"},
			{"1970-10-10", "1970-10-10T00:00:00.000Z"},
			{"1970-11-11", "1970-11-11T00:00:00.000Z"},

			// date and time only YYYY-MM-DDThh:mm:ssZ
			{"1970-01-01T01:01:01Z", "1970-01-01T01:01:01.000Z"},
			{"1970-01-01T10:10:10Z", "1970-01-01T10:10:10.000Z"},
			{"1970-01-01T11:11:11Z", "1970-01-01T11:11:11.000Z"},

			// date and time + milli only YYYY-MM-DDThh:mm:ss.SSSZ
			{"1970-01-01T00:00:00.000Z", "1970-01-01T00:00:00.000Z"},
			{"1970-01-01T00:00:00.001Z", "1970-01-01T00:00:00.001Z"},
			{"1970-01-01T00:00:00.011Z", "1970-01-01T00:00:00.011Z"},
			{"1970-01-01T00:00:00.111Z", "1970-01-01T00:00:00.111Z"},
			{"1970-01-01T00:00:00.010Z", "1970-01-01T00:00:00.010Z"},
			{"1970-01-01T00:00:00.100Z", "1970-01-01T00:00:00.100Z"},

			{"1970-01-01T00:00:00+0001", "1970-01-01T00:01:00.000Z"},
			{"1970-01-01T00:00:00+0010", "1970-01-01T00:10:00.000Z"},
			{"1970-01-01T00:00:00+0100", "1970-01-01T01:00:00.000Z"},
			{"1970-01-01T00:00:00+1000", "1970-01-01T10:00:00.000Z"},

			{"1970-01-01T00:01:00-0001", "1970-01-01T00:00:00.000Z"},
			{"1970-01-01T00:10:00-0010", "1970-01-01T00:00:00.000Z"},
			{"1970-01-01T01:00:00-0100", "1970-01-01T00:00:00.000Z"},
			{"1970-01-01T10:00:00-1000", "1970-01-01T00:00:00.000Z"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.out), func(t *testing.T) {
				t.Parallel()
				d, err := types.ParseDatetime(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.out)
			})
		}
	}

	{
		tests := []struct{ in, errStr string }{
			{"", "error parsing datetime value: string too short"},
			{"-", "error parsing datetime value: string too short"},
			{"012345678", "error parsing datetime value: string too short"},

			{"195-01-01T00:00:00Z", "error parsing datetime value: invalid year"},
			{"1995-1-01T00:00:00Z", "error parsing datetime value: invalid month"},
			{"1995-01-0T00:00:00Z", "error parsing datetime value: invalid day"},
			{"1995-01T00:00:00Z", "error parsing datetime value: unexpected character 'T'"},
			{"1995-01-01T:00:00Z", "error parsing datetime value: invalid time"},
			{"1995-01-01Taa:00:00Z", "error parsing datetime value: invalid hour"},
			{"1995-01-01T00:aa:00Z", "error parsing datetime value: invalid minute"},
			{"1995-01-01T00:00:aaZ", "error parsing datetime value: invalid second"},
			{"1995-01-01T00:00:00Zgarbage", "error parsing datetime value: unexpected trailer after timezone indicator"},
			{"1995-01-01T00:00:00.", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.0", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.00", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.aaa", "error parsing datetime value: invalid millisecond"},

			{"1995-01-01T00:00:00.001", "error parsing datetime value: expected timezone indicator"},

			{"1995-01-01T00:00:00.000Z+", "error parsing datetime value: unexpected trailer after timezone indicator"},
			{"1995-01-01T00:00:00.000Z+0000", "error parsing datetime value: unexpected trailer after timezone indicator"},
			{"1995-01-01T00:00:00.000Z+000", "error parsing datetime value: unexpected trailer after timezone indicator"},

			{"1995-01-01T00:00:00.000+", "error parsing datetime value: expected time offset"},

			{"1995-01-01T00:00:00.000+", "error parsing datetime value: expected time offset"},
			{"1995-01-01T00:00:00.000-", "error parsing datetime value: expected time offset"},

			{"1995-01-01T00:00:00.000-0", "error parsing datetime value: expected time offset"},
			{"1995-01-01T00:00:00.000-00", "error parsing datetime value: expected time offset"},
			{"1995-01-01T00:00:00.000-000", "error parsing datetime value: expected time offset"},
			{"1995-01-01T00:00:00.000-000a", "error parsing datetime value: invalid time offset"},
			{"1995-01-01T00:00:00.000-00aa", "error parsing datetime value: invalid time offset"},
			{"1995-01-01T00:00:00.000-0aaa", "error parsing datetime value: invalid time offset"},
			{"1995-01-01T00:00:00.000-aaaa", "error parsing datetime value: invalid time offset"},
			{"1995-01-01T00:00:00.000-aaaa0", "error parsing datetime value: unexpected trailer"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDatetime(tt.in)
				testutil.ErrorIs(t, err, types.ErrDatetime)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.UnsafeDatetime(1)
		one2 := types.UnsafeDatetime(1)
		zero := types.UnsafeDatetime(0)
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.UnsafeDatetime(42).MarshalCedar()), `datetime("1970-01-01T00:00:00.042Z")`)
	})

}
