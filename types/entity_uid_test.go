package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestEntity(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		twoElems := types.EntityUID{"type", "id"}
		twoElems2 := types.EntityUID{"type", "id"}
		differentValues := types.EntityUID{"asdf", "vfds"}
		testutil.FatalIf(t, !twoElems.Equal(twoElems), "%v not Equal to %v", twoElems, twoElems)
		testutil.FatalIf(t, !twoElems.Equal(twoElems2), "%v not Equal to %v", twoElems, twoElems2)
		testutil.FatalIf(t, twoElems.Equal(differentValues), "%v Equal to %v", twoElems, differentValues)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.EntityUID{Type: "type", ID: "id"}.String(), `type::"id"`)
		testutil.Equals(t, types.EntityUID{Type: "namespace::type", ID: "id"}.String(), `namespace::type::"id"`)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.EntityUID{"type", "id"}.MarshalCedar()), `type::"id"`)
	})
}

func TestEntityUIDSet(t *testing.T) {
	t.Parallel()

	t.Run("new empty set", func(t *testing.T) {
		emptySets := []types.EntityUIDSet{
			types.NewEntityUIDSet(),
			types.NewEntityUIDSet(0),
			types.NewEntityUIDSet(1),
			types.NewEntityUIDSetFromSlice(nil),
			types.NewEntityUIDSetFromSlice([]types.EntityUID{}),
		}

		for _, es := range emptySets {
			testutil.Equals(t, es.Len(), 0)
			testutil.Equals(t, emptySets[0].Equal(es), true)
			testutil.Equals(t, es.Equal(emptySets[0]), true)
		}
	})

	t.Run("new set from slice", func(t *testing.T) {
		a := types.NewEntityUID("typeA", "1")
		b := types.NewEntityUID("typeB", "2")
		o := types.NewEntityUID("typeO", "2")
		s1 := types.NewEntityUIDSet()
		s1.Add(a)
		s1.Add(b)
		s1.Add(o)

		s2 := types.NewEntityUIDSetFromSlice([]types.EntityUID{o, b, a})

		testutil.Equals(t, s1.Equal(s2), true)
	})
}
