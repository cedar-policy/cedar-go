package types_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal"
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

			{"1972-02-29T10:00:00-1000", "1972-02-29T00:00:00.000Z"},
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
			{"1995+01-01T00:00:00Z", "error parsing datetime value: unexpected character '+'"},
			{"1995-01+01T00:00:00Z", "error parsing datetime value: unexpected character '+'"},
			{"1995-01-01T00+00:00Z", "error parsing datetime value: unexpected character '+'"},
			{"1995-01-01T00:00+00Z", "error parsing datetime value: unexpected character '+'"},
			{"1995-01-00Y00:00:00Z", "error parsing datetime value: unexpected character 'Y'"},
			{"1995-01-00T00:00:00V", "error parsing datetime value: invalid time zone designator"},

			{"1995-1-01T00:00:00Z", "error parsing datetime value: invalid month"},
			{"1995-01-0T00:00:00Z", "error parsing datetime value: invalid day"},
			{"1995-01T00:00:00Z", "error parsing datetime value: unexpected character 'T'"},
			{"1995-01-01T:00:00Z", "error parsing datetime value: invalid time"},
			{"1995-01-01Taa:00:00Z", "error parsing datetime value: invalid hour"},
			{"1995-01-01T00:aa:00Z", "error parsing datetime value: invalid minute"},
			{"1995-01-01T00:00:aaZ", "error parsing datetime value: invalid second"},
			{"1995-01-01T00:00:00Zgarbage", "error parsing datetime value: unexpected trailer after time zone designator"},
			{"1995-01-01T00:00:00.", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.0", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.00", "error parsing datetime value: invalid millisecond"},
			{"1995-01-01T00:00:00.aaa", "error parsing datetime value: invalid millisecond"},

			{"1995-01-01T00:00:00.001", "error parsing datetime value: expected time zone designator"},

			{"1995-01-01T00:00:00.000Z+", "error parsing datetime value: unexpected trailer after time zone designator"},
			{"1995-01-01T00:00:00.000Z+0000", "error parsing datetime value: unexpected trailer after time zone designator"},
			{"1995-01-01T00:00:00.000Z+000", "error parsing datetime value: unexpected trailer after time zone designator"},

			{"1995-01-01T00:00:00.000+", "error parsing datetime value: invalid time zone offset"},

			{"1995-01-01T00:00:00.000+", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-", "error parsing datetime value: invalid time zone offset"},

			{"1995-01-01T00:00:00.000-0", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-00", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-000", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-000a", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-00aa", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-0aaa", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-aaaa", "error parsing datetime value: invalid time zone offset"},
			{"1995-01-01T00:00:00.000-aaaa0", "error parsing datetime value: unexpected trailer after time zone designator"},

			{"1995-04-31T00:00:00Z", "error parsing datetime value: invalid date"},

			// Prevent Wrapping invalid dates to real dates: See: cedar-policy/rfcs#94
			{"2024-02-30T00:00:00Z", "error parsing datetime value: invalid date"},
			{"2024-02-29T23:59:60Z", "error parsing datetime value: second is out of range"},
			{"2023-02-28T23:59:60Z", "error parsing datetime value: second is out of range"},
			{"2023-02-28T23:60:59Z", "error parsing datetime value: minute is out of range"},
			{"1970-01-01T25:00:00Z", "error parsing datetime value: hour is out of range"},
			{"1970-12-32T:00:00Z", "error parsing datetime value: day is out of range"},
			{"1970-13-01T00:00:00Z", "error parsing datetime value: month is out of range"},

			{"1970-01-01T00:00:00+2400", "error parsing datetime value: time zone offset hours are out of range"},
			{"1970-01-01T00:00:00-2400", "error parsing datetime value: time zone offset hours are out of range"},
			{"1970-01-01T00:00:00+2360", "error parsing datetime value: time zone offset minutes are out of range"},
			{"1970-01-01T00:00:00-2360", "error parsing datetime value: time zone offset minutes are out of range"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDatetime(tt.in)
				testutil.ErrorIs(t, err, internal.ErrDatetime)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("Construct", func(t *testing.T) {
		t.Parallel()
		one := types.NewDatetimeFromMillis(1)
		two := types.NewDatetime(time.UnixMilli(1))
		testutil.Equals(t, one.Milliseconds(), two.Milliseconds())
	})

	t.Run("Time", func(t *testing.T) {
		t.Parallel()
		in := types.NewDatetime(time.UnixMilli(42))
		got := in.Time()
		want := time.UnixMilli(42).UTC()
		testutil.Equals(t, got, want)
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.NewDatetimeFromMillis(1)
		one2 := types.NewDatetime(time.UnixMilli(1))
		zero := types.NewDatetime(time.UnixMilli(0))
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("LessThan", func(t *testing.T) {
		t.Parallel()
		one := types.NewDatetime(time.UnixMilli(1))
		zero := types.NewDatetime(time.UnixMilli(0))
		f := types.Boolean(false)

		tests := []struct {
			l       types.Datetime
			r       types.Value
			want    bool
			wantErr error
		}{
			{one, zero, false, nil},
			{zero, one, true, nil},
			{zero, zero, false, nil},
			{zero, f, false, internal.ErrNotComparable},
		}

		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("LessThan_%d_%v<%v", ti, tt.l, tt.r), func(t *testing.T) {
				t.Parallel()
				got, gotErr := tt.l.LessThan(tt.r)
				testutil.Equals(t, got, tt.want)
				testutil.ErrorIs(t, gotErr, tt.wantErr)
			})
		}

	})

	t.Run("LessThanOrEqual", func(t *testing.T) {
		t.Parallel()
		one := types.NewDatetime(time.UnixMilli(1))
		zero := types.NewDatetime(time.UnixMilli(0))
		f := types.Boolean(false)

		tests := []struct {
			l       types.Datetime
			r       types.Value
			want    bool
			wantErr error
		}{
			{one, zero, false, nil},
			{zero, one, true, nil},
			{zero, zero, true, nil},
			{zero, f, false, internal.ErrNotComparable},
		}

		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("LessThanOrEqual_%d_%v<%v", ti, tt.l, tt.r), func(t *testing.T) {
				t.Parallel()
				got, gotErr := tt.l.LessThanOrEqual(tt.r)
				testutil.Equals(t, got, tt.want)
				testutil.ErrorIs(t, gotErr, tt.wantErr)
			})
		}
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.NewDatetime(time.UnixMilli(42)).MarshalCedar()), `datetime("1970-01-01T00:00:00.042Z")`)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		expected := `{
			"__extn": {
				"fn": "datetime",
				"arg": "1970-01-01T00:00:00.042Z"
			}
		}`
		dt1 := types.NewDatetime(time.UnixMilli(42))
		testutil.JSONMarshalsTo(t, dt1, expected)

		var dt2 types.Datetime
		err := json.Unmarshal([]byte(expected), &dt2)
		testutil.OK(t, err)
		testutil.Equals(t, dt1, dt2)
	})

	t.Run("UnmarshalJSON/error", func(t *testing.T) {
		t.Parallel()
		var dt2 types.Datetime
		err := json.Unmarshal([]byte("{}"), &dt2)
		testutil.Error(t, err)
	})
}
