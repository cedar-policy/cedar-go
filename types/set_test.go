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

		testutil.FatalIf(t, !empty.Equal(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equal(empty2), "%v not Equal to %v", empty, empty2)
		testutil.FatalIf(t, !oneTrue.Equal(oneTrue), "%v not Equal to %v", oneTrue, oneTrue)
		testutil.FatalIf(t, !oneTrue.Equal(oneTrue2), "%v not Equal to %v", oneTrue, oneTrue2)
		testutil.FatalIf(t, !nestedOnce.Equal(nestedOnce), "%v not Equal to %v", nestedOnce, nestedOnce)
		testutil.FatalIf(t, !nestedOnce.Equal(nestedOnce2), "%v not Equal to %v", nestedOnce, nestedOnce2)
		testutil.FatalIf(t, !nestedTwice.Equal(nestedTwice), "%v not Equal to %v", nestedTwice, nestedTwice)
		testutil.FatalIf(t, !nestedTwice.Equal(nestedTwice2), "%v not Equal to %v", nestedTwice, nestedTwice2)
		testutil.FatalIf(t, !oneTwoThree.Equal(threeTwoTwoOne), "%v not Equal to %v", oneTwoThree, threeTwoTwoOne)

		testutil.FatalIf(t, empty.Equal(oneFalse), "%v Equal to %v", empty, oneFalse)
		testutil.FatalIf(t, oneTrue.Equal(oneFalse), "%v Equal to %v", oneTrue, oneFalse)
		testutil.FatalIf(t, nestedOnce.Equal(nestedTwice), "%v Equal to %v", nestedOnce, nestedTwice)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Set{}.String(), "[]")
		testutil.Equals(
			t,
			types.Set{types.Boolean(true), types.Long(1)}.String(),
			"[true, 1]")
	})

}
