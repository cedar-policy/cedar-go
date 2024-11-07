package types_test

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestLong(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.Long(1)
		one2 := types.Long(1)
		zero := types.Long(0)
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("LessThan", func(t *testing.T) {
		t.Parallel()
		one := types.Long(1)
		zero := types.Long(0)
		f := types.Boolean(false)

		tests := []struct {
			l       types.Long
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
		one := types.Long(1)
		zero := types.Long(0)
		f := types.Boolean(false)

		tests := []struct {
			l       types.Long
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

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Long(1).String(), "1")
	})
}
