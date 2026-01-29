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

			{"1970-01-01T00:00:00+0001", "1969-12-31T23:59:00.000Z"},
			{"1970-01-01T00:00:00+0010", "1969-12-31T23:50:00.000Z"},
			{"1970-01-01T00:00:00+0100", "1969-12-31T23:00:00.000Z"},
			{"1970-01-01T00:00:00+1000", "1969-12-31T14:00:00.000Z"},

			{"1970-01-01T00:00:00-0001", "1970-01-01T00:01:00.000Z"},
			{"1970-01-01T00:00:00-0010", "1970-01-01T00:10:00.000Z"},
			{"1970-01-01T00:00:00-0100", "1970-01-01T01:00:00.000Z"},
			{"1970-01-01T00:00:00-1000", "1970-01-01T10:00:00.000Z"},

			{"1972-02-29T10:00:00+1000", "1972-02-29T00:00:00.000Z"},

			// Expanded year format (RFC 110)
			{"+000000010-01-01", "0010-01-01T00:00:00.000Z"},
			{"+000001970-06-15", "1970-06-15T00:00:00.000Z"},
			{"+000009999-12-31", "9999-12-31T00:00:00.000Z"},
			{"+000010000-01-01", "+000010000-01-01T00:00:00.000Z"},
			{"+000100000-06-15", "+000100000-06-15T00:00:00.000Z"},
			{"+001000000-12-31", "+001000000-12-31T00:00:00.000Z"},
			{"-000000001-01-01", "-000000001-01-01T00:00:00.000Z"},
			{"-000001000-06-15", "-000001000-06-15T00:00:00.000Z"},
			{"-000010000-12-31", "-000010000-12-31T00:00:00.000Z"},
			{"+000010000-01-01T12:30:45.123Z", "+000010000-01-01T12:30:45.123Z"},
			{"-000000100-01-01T00:00:00.001Z", "-000000100-01-01T00:00:00.001Z"},
			{"+292278994-08-17T07:12:55.807Z", "+292278994-08-17T07:12:55.807Z"},
			{"+292278994-08-17T06:12:55.807-0100", "+292278994-08-17T07:12:55.807Z"},
			{"+292278994-08-17T08:12:55.807+0100", "+292278994-08-17T07:12:55.807Z"},
			{"-292275055-05-17T16:47:04.192Z", "-292275055-05-17T16:47:04.192Z"},
			{"-292275055-05-17T15:47:04.192-0100", "-292275055-05-17T16:47:04.192Z"},
			{"-292275055-05-17T17:47:04.192+0100", "-292275055-05-17T16:47:04.192Z"},
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
			{"", "error parsing datetime value: unexpected EOF"},
			{"*", "error parsing datetime value: invalid year"},
			{"012345678", "error parsing datetime value: unexpected character 4"},

			{"195-01-01T00:00:00Z", "error parsing datetime value: invalid year"},
			{"1995+01-01T00:00:00Z", "error parsing datetime value: unexpected character +"},
			{"1995-01+01T00:00:00Z", "error parsing datetime value: unexpected character +"},
			{"1995-01-01T00+00:00Z", "error parsing datetime value: unexpected character +"},
			{"1995-01-01T00:00+00Z", "error parsing datetime value: unexpected character +"},
			{"1995-01-01Y00:00:00Z", "error parsing datetime value: unexpected character Y"},
			{"1995-01-01T00:00:00V", "error parsing datetime value: invalid time zone designator"},

			{"1995-1-01T00:00:00Z", "error parsing datetime value: invalid month"},
			{"1995-01-0T00:00:00Z", "error parsing datetime value: invalid day"},
			{"1995-01", "error parsing datetime value: unexpected EOF"},
			{"1995-01T00:00:00Z", "error parsing datetime value: unexpected character T"},
			{"1995-01-01T:00:00Z", "error parsing datetime value: invalid hour"},
			{"1995-01-01Taa:00:00Z", "error parsing datetime value: invalid hour"},
			{"1995-01-01T00:aa:00Z", "error parsing datetime value: invalid minute"},
			{"1995-01-01T00:00:aaZ", "error parsing datetime value: invalid second"},
			{"1995-01-01T00:00:00", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.0", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.00", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.aaa", "error parsing datetime value: invalid millisecond"},

			{"1995-01-01T00:00:00.001", "error parsing datetime value: unexpected EOF"},

			{"1995-01-01T00:00:00.000+", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.000-", "error parsing datetime value: unexpected EOF"},

			{"1995-01-01T00:00:00.000-0", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.000-00", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.000-000", "error parsing datetime value: unexpected EOF"},
			{"1995-01-01T00:00:00.000-000a", "error parsing datetime value: invalid offset minutes"},
			{"1995-01-01T00:00:00.000-00aa", "error parsing datetime value: invalid offset minutes"},
			{"1995-01-01T00:00:00.000-0aaa", "error parsing datetime value: invalid offset hours"},
			{"1995-01-01T00:00:00.000-aaaa", "error parsing datetime value: invalid offset hours"},

			{"1995-04-31T00:00:00Z", "error parsing datetime value: invalid date"},

			// Prevent Wrapping invalid dates to real dates: See: cedar-policy/rfcs#94
			{"2024-02-30T00:00:00Z", "error parsing datetime value: invalid date"},
			{"2024-02-29T23:59:60Z", "error parsing datetime value: second is greater than 59"},
			{"2023-02-28T23:59:60Z", "error parsing datetime value: second is greater than 59"},
			{"2023-02-28T23:60:59Z", "error parsing datetime value: minute is greater than 59"},
			{"1970-01-01T25:00:00Z", "error parsing datetime value: hour is greater than 23"},
			{"1970-12-32T:00:00Z", "error parsing datetime value: day is greater than 31"},
			{"1970-13-01T00:00:00Z", "error parsing datetime value: month is greater than 12"},

			{"1970-01-01T00:00:00+2400", "error parsing datetime value: offset hours is greater than 23"},
			{"1970-01-01T00:00:00-2400", "error parsing datetime value: offset hours is greater than 23"},
			{"1970-01-01T00:00:00+2360", "error parsing datetime value: offset minutes is greater than 59"},
			{"1970-01-01T00:00:00-2360", "error parsing datetime value: offset minutes is greater than 59"},

			{"+", "error parsing datetime value: unexpected EOF"},
			{"-", "error parsing datetime value: unexpected EOF"},
			{"+12345678", "error parsing datetime value: unexpected EOF"},
			{"+1234-01-01", "error parsing datetime value: invalid year"},
			{"+00000000a-01-01", "error parsing datetime value: invalid year"},
			{"-abcdefghi-01-01", "error parsing datetime value: invalid year"},
			{"+12345678A-01-01", "error parsing datetime value: invalid year"},

			{"1972-02-29T10:00:00-1000x", "error parsing datetime value: unexpected additional characters"},

			{"+292278994-08-17T07:12:55.808Z", "error parsing datetime value: timestamp out of range"},
			{"+292278994-08-17T06:12:55.808-0100", "error parsing datetime value: timestamp out of range"},
			{"+292278994-08-17T08:12:55.808+0100", "error parsing datetime value: timestamp out of range"},
			{"-292275055-05-17T16:47:04.191Z", "error parsing datetime value: timestamp out of range"},
			{"-292275055-05-17T15:47:04.191-0100", "error parsing datetime value: timestamp out of range"},
			{"-292275055-05-17T17:47:04.191+0100", "error parsing datetime value: timestamp out of range"},
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
		tests := []struct {
			year     int
			month    int
			day      int
			expected string
		}{
			{0, 1, 1, `datetime("0000-01-01T00:00:00.000Z")`},
			{1970, 1, 1, `datetime("1970-01-01T00:00:00.000Z")`},
			{9999, 12, 31, `datetime("9999-12-31T00:00:00.000Z")`},
			{10000, 1, 1, `datetime("+000010000-01-01T00:00:00.000Z")`},
			{100000, 6, 15, `datetime("+000100000-06-15T00:00:00.000Z")`},
			{1000000, 12, 31, `datetime("+001000000-12-31T00:00:00.000Z")`},
			{-1, 1, 1, `datetime("-000000001-01-01T00:00:00.000Z")`},
			{-100, 6, 15, `datetime("-000000100-06-15T00:00:00.000Z")`},
			{-10000, 12, 31, `datetime("-000010000-12-31T00:00:00.000Z")`},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_year=%d", ti, tt.year), func(t *testing.T) {
				t.Parallel()
				dt := types.NewDatetime(time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0, time.UTC))
				testutil.Equals(t, dt.MarshalCedar(), []byte(tt.expected))
			})
		}
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		t.Run("FourDigitYear", func(t *testing.T) {
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

		t.Run("FiveDigitYear", func(t *testing.T) {
			t.Parallel()
			dt := types.NewDatetime(time.Date(10000, 1, 1, 0, 0, 0, 42000000, time.UTC))
			expected := `{
				"__extn": {
					"fn": "datetime",
					"arg": "+000010000-01-01T00:00:00.042Z"
				}
			}`
			testutil.JSONMarshalsTo(t, dt, expected)

			var dt2 types.Datetime
			err := json.Unmarshal([]byte(expected), &dt2)
			testutil.OK(t, err)
			testutil.Equals(t, dt, dt2)
		})
	})

	t.Run("UnmarshalJSON/error", func(t *testing.T) {
		t.Parallel()
		var dt2 types.Datetime
		err := json.Unmarshal([]byte("{}"), &dt2)
		testutil.Error(t, err)
	})

}
