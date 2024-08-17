package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := types.Set{}
		empty2 := types.Set{}
		oneTrue := types.Set{types.Boolean(true)}
		oneTrue2 := types.Set{types.Boolean(true)}
		oneFalse := types.Set{types.Boolean(false)}
		nestedOnce := types.Set{empty, oneTrue, oneFalse}
		nestedOnce2 := types.Set{empty, oneTrue, oneFalse}
		nestedTwice := types.Set{empty, oneTrue, oneFalse, nestedOnce}
		nestedTwice2 := types.Set{empty, oneTrue, oneFalse, nestedOnce}
		oneTwoThree := types.Set{
			types.Long(1), types.Long(2), types.Long(3),
		}
		threeTwoTwoOne := types.Set{
			types.Long(3), types.Long(2), types.Long(2), types.Long(1),
		}

		testutil.FatalIf(t, !empty.Equals(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equals(empty2), "%v not Equal to %v", empty, empty2)
		testutil.FatalIf(t, !oneTrue.Equals(oneTrue), "%v not Equal to %v", oneTrue, oneTrue)
		testutil.FatalIf(t, !oneTrue.Equals(oneTrue2), "%v not Equal to %v", oneTrue, oneTrue2)
		testutil.FatalIf(t, !nestedOnce.Equals(nestedOnce), "%v not Equal to %v", nestedOnce, nestedOnce)
		testutil.FatalIf(t, !nestedOnce.Equals(nestedOnce2), "%v not Equal to %v", nestedOnce, nestedOnce2)
		testutil.FatalIf(t, !nestedTwice.Equals(nestedTwice), "%v not Equal to %v", nestedTwice, nestedTwice)
		testutil.FatalIf(t, !nestedTwice.Equals(nestedTwice2), "%v not Equal to %v", nestedTwice, nestedTwice2)
		testutil.FatalIf(t, !oneTwoThree.Equals(threeTwoTwoOne), "%v not Equal to %v", oneTwoThree, threeTwoTwoOne)

		testutil.FatalIf(t, empty.Equals(oneFalse), "%v Equal to %v", empty, oneFalse)
		testutil.FatalIf(t, oneTrue.Equals(oneFalse), "%v Equal to %v", oneTrue, oneFalse)
		testutil.FatalIf(t, nestedOnce.Equals(nestedTwice), "%v Equal to %v", nestedOnce, nestedTwice)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		AssertValueString(t, types.Set{}, "[]")
		AssertValueString(
			t,
			types.Set{types.Boolean(true), types.Long(1)},
			"[true, 1]")
	})

	t.Run("TypeName", func(t *testing.T) {
		t.Parallel()
		tn := types.Set{}.TypeName()
		testutil.Equals(t, tn, "set")
	})
}
