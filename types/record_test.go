package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestRecord(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		empty := types.Record{}
		empty2 := types.Record{}
		twoElems := types.Record{
			"foo": types.Boolean(true),
			"bar": types.String("blah"),
		}
		twoElems2 := types.Record{
			"foo": types.Boolean(true),
			"bar": types.String("blah"),
		}
		differentValues := types.Record{
			"foo": types.Boolean(false),
			"bar": types.String("blaz"),
		}
		differentKeys := types.Record{
			"foo": types.Boolean(false),
			"bar": types.Long(1),
		}
		nested := types.Record{
			"one":  types.Long(1),
			"two":  types.Long(2),
			"nest": twoElems,
		}
		nested2 := types.Record{
			"one":  types.Long(1),
			"two":  types.Long(2),
			"nest": twoElems,
		}

		testutil.FatalIf(t, !empty.Equal(empty), "%v not Equal to %v", empty, empty)
		testutil.FatalIf(t, !empty.Equal(empty2), "%v not Equal to %v", empty, empty2)

		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)

		testutil.FatalIf(t, !nested.Equal(nested), "%v not Equal to %v", nested, nested)
		testutil.FatalIf(t, !nested.Equal(nested2), "%v not Equal to %v", nested, nested2)

		testutil.FatalIf(t, nested.Equal(twoElems), "%v Equal to %v", nested, twoElems)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
		testutil.FatalIf(t, twoElems.Equal(differentKeys), "%v Equal to %v", twoElems, differentKeys)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		assertValueString(t, types.Record{}, "{}")
		assertValueString(
			t,
			types.Record{"foo": types.Boolean(true)},
			`{"foo":true}`)
		assertValueString(
			t,
			types.Record{
				"foo": types.Boolean(true),
				"bar": types.String("blah"),
			},
			`{"bar":"blah", "foo":true}`)
	})

}
