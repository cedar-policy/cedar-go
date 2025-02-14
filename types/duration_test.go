package types_test

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/internal"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestDuration(t *testing.T) {
	t.Parallel()
	{
		tests := []struct{ in, out string }{
			{"1h", "1h"},
			{"60m", "1h"},
			{"3600s", "1h"},
			{"3600000ms", "1h"},
			{"24h", "1d"},
			{"36h", "1d12h"},
			{"1d12h", "1d12h"},
			{"1d11h60m", "1d12h"},
			{"1d11h59m60s", "1d12h"},
			{"1d11h59m59s1000ms", "1d12h"},
			{"60s60000ms", "2m"},
			{"62m", "1h2m"},
			{"2m3600s", "1h2m"},
			{"-1h", "-1h"},
			{"-60m", "-1h"},
			{"-3600s", "-1h"},
			{"-3600000ms", "-1h"},
			{"-24h", "-1d"},
			{"-36h", "-1d12h"},
			{"-1d12h", "-1d12h"},
			{"-1d11h60m", "-1d12h"},
			{"-1d11h59m60s", "-1d12h"},
			{"-1d11h59m59s1000ms", "-1d12h"},
			{"-60s60000ms", "-2m"},
			{"-62m", "-1h2m"},
			{"-2m3600s", "-1h2m"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.out), func(t *testing.T) {
				t.Parallel()
				d, err := types.ParseDuration(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.out)
			})
		}
	}

	{
		tests := []struct{ in, errStr string }{
			{"", "error parsing duration value: string too short"},
			{"-", "error parsing duration value: string too short"},
			{"h", "error parsing duration value: string too short"},
			{"3", "error parsing duration value: string too short"},
			{"-m", "error parsing duration value: unit found without quantity"},
			{"-1t", "error parsing duration value: unexpected character 't'"},
			{"-1h1h", "error parsing duration value: unexpected unit 'h'"},
			{"-3h3", "error parsing duration value: expected unit"},
			{"3h-1m", "error parsing duration value: unexpected character '-'"},
			{"3h1m   ", "error parsing duration value: unexpected character ' '"},
			{"3600ms30ms", "error parsing duration value: invalid duration"},
			{"36ms30h", "error parsing duration value: invalid duration"},
			{"999999999999999999999ms", "error parsing duration value: overflow"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDuration(tt.in)
				testutil.ErrorIs(t, err, internal.ErrDuration)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("Construct", func(t *testing.T) {
		t.Parallel()
		one := types.NewDurationFromMillis(1)
		two := types.NewDuration(1 * time.Millisecond)
		testutil.Equals(t, one.ToMilliseconds(), two.ToMilliseconds())
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name string
			in   types.Duration
			out  time.Duration
			err  func(testutil.TB, error)
		}{
			{"ok", types.NewDuration(time.Millisecond * 42), time.Millisecond * 42, testutil.OK},
			{"maxPlusOne", types.NewDurationFromMillis(math.MaxInt64/1000 + 1), 0, testutil.Error},
			{"minMinusOne", types.NewDurationFromMillis(math.MinInt64/1000 - 1), 0, testutil.Error},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				out, err := tt.in.Duration()
				testutil.Equals(t, out, tt.out)
				tt.err(t, err)
			})
		}
	})

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.NewDurationFromMillis(1)
		one2 := types.NewDuration(1 * time.Millisecond)
		zero := types.NewDuration(time.Duration(0))
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("LessThan", func(t *testing.T) {
		t.Parallel()
		one := types.NewDuration(1 * time.Millisecond)
		zero := types.NewDuration(time.Duration(0))
		f := types.Boolean(false)

		tests := []struct {
			l       types.Duration
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
		one := types.NewDuration(1 * time.Millisecond)
		zero := types.NewDuration(time.Duration(0))
		f := types.Boolean(false)

		tests := []struct {
			l       types.Duration
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

	t.Run("ToUnit", func(t *testing.T) {
		t.Parallel()
		dur := types.NewDuration((26 * time.Hour) + (31 * time.Minute) + (43 * time.Second) + (17 * time.Millisecond))

		testutil.Equals(t, dur.ToDays(), 1)
		testutil.Equals(t, dur.ToHours(), 26)
		testutil.Equals(t, dur.ToMinutes(), 1591)
		testutil.Equals(t, dur.ToSeconds(), 95503)
		testutil.Equals(t, dur.ToMilliseconds(), 95503017)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.NewDuration(42*time.Millisecond).MarshalCedar()), `duration("42ms")`)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		expected := `{
			"__extn": {
				"fn": "duration",
				"arg": "42ms"
			}
		}`
		d1 := types.NewDuration(42 * time.Millisecond)
		testutil.JSONMarshalsTo(t, d1, expected)

		var d2 types.Duration
		err := json.Unmarshal([]byte(expected), &d2)
		testutil.OK(t, err)
		testutil.Equals(t, d1, d2)
	})

	t.Run("UnmarshalJSON/error", func(t *testing.T) {
		t.Parallel()
		var dt2 types.Duration
		err := json.Unmarshal([]byte("{}"), &dt2)
		testutil.Error(t, err)
	})
}
