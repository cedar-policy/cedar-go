package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestBool(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		t1 := types.Boolean(true)
		t2 := types.Boolean(true)
		f := types.Boolean(false)
		zero := types.Long(0)
		testutil.FatalIf(t, !t1.Equal(t1), "%v not Equal to %v", t1, t1)
		testutil.FatalIf(t, !t1.Equal(t2), "%v not Equal to %v", t1, t2)
		testutil.FatalIf(t, t1.Equal(f), "%v Equal to %v", t1, f)
		testutil.FatalIf(t, f.Equal(t1), "%v Equal to %v", f, t1)
		testutil.FatalIf(t, f.Equal(zero), "%v Equal to %v", f, zero)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Boolean(true).String(), "true")
		testutil.Equals(t, types.Boolean(false).String(), "false")
	})
}
